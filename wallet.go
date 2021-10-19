package cardano

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/echovl/cardano-go/crypto"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/tyler-smith/go-bip39"
	"errors"
)

const (
	entropySizeInBits         = 160
	purposeIndex       uint32 = 1852 + 0x80000000
	coinTypeIndex      uint32 = 1815 + 0x80000000
	accountIndex       uint32 = 0x80000000
	externalChainIndex uint32 = 0x0
	walleIDAlphabet           = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

type Wallet struct {
	ID      string
	Name    string
	Keys 	map[string]crypto.ExtendedSigningKey
	Skeys   []crypto.ExtendedSigningKey
	pkeys   []crypto.ExtendedVerificationKey
	rootKey crypto.ExtendedSigningKey
	node    cardanoNode
	network Network
	Utxos []Utxo
	Tip NodeTip
}

func (w *Wallet) SetNetwork(net Network) {
	w.network = net
}
func (w *Wallet) SetKey(key string) error{
	key = strings.Replace(key,"0x","",1)
	prikey ,err := hex.DecodeString(key)
	if err != nil{
		return err
	}
	pri := crypto.ExtendedSigningKey(prikey)
	addr := newEnterpriseAddress(pri.ExtendedVerificationKey(), w.network)
	if w.Keys == nil {
		w.Keys = make(map[string]crypto.ExtendedSigningKey)
	}
	w.Keys[string(addr)] = pri
	return nil
}

// Transfer sends an amount of lovelace to the receiver address
//TODO: remove hardcoded protocol parameters, these parameters must be obtained using the cardano node
func (w *Wallet) Transfer(receiver Address, amount uint64,changeAddress Address) (tx *transaction,err error) {
	// Calculate if the account has enough balance
	balance, err := w.Balance()
	if err != nil {
		return tx,err
	}
	if amount > balance {
		return tx,fmt.Errorf("Not enough balance, %v > %v", amount, balance)
	}

	// Find utxos that cover the amount to transfer
	pickedUtxos := []Utxo{}
	utxos, err := w.findUtxos()
	pickedUtxosAmount := uint64(0)
	for _, utxo := range utxos {
		if pickedUtxosAmount > amount {
			break
		}
		pickedUtxos = append(pickedUtxos, utxo)
		pickedUtxosAmount += utxo.Amount
	}

	builder := newTxBuilder(protocolParams{
		MinimumUtxoValue: 1000000,
		MinFeeA:          44,
		MinFeeB:          155381,
	})

	keys := make(map[int]crypto.ExtendedSigningKey)
	for i, utxo := range pickedUtxos {
		for _, key := range w.Keys {
			vkey := key.ExtendedVerificationKey()
			address := newEnterpriseAddress(vkey, w.network)
			if address == utxo.Address {
				keys[i] = key
			}
		}
	}

	if len(keys) != len(pickedUtxos) {
		panic("not enough keys")
	}

	for i, utxo := range pickedUtxos {
		skey := keys[i]
		vkey := skey.ExtendedVerificationKey()
		builder.AddInput(vkey, utxo.TxId, utxo.Index, utxo.Amount)
	}
	builder.AddOutput(receiver, amount)

	// Calculate and set ttl
	tip := w.Tip
	builder.SetTtl(tip.Slot + 1200)
	if changeAddress == "" {
		changeAddress = pickedUtxos[0].Address
	}

	err = builder.AddFee(changeAddress)
	if err != nil {
		return tx,err
	}
	for _, key := range keys {
		builder.Sign(key)
	}
	tx2 := builder.Build()
	return &tx2,nil
}

// Balance returns the total lovelace amount of the wallet.
func (w *Wallet) Balance() (uint64, error) {
	var balance uint64
	utxos, err := w.findUtxos()
	if err != nil {
		return 0, nil
	}
	for _, utxo := range utxos {
		balance += utxo.Amount
	}
	return balance, nil
}

func (w *Wallet) findUtxos() ([]Utxo, error) {
	return w.Utxos,nil
	//addresses := w.Addresses()
	//walletUtxos := []Utxo{}
	//for _, addr := range addresses {
	//	addrUtxos, err := w.node.QueryUtxos(addr)
	//	if err != nil {
	//		return nil, err
	//	}
	//	walletUtxos = append(walletUtxos, addrUtxos...)
	//}
	//return walletUtxos, nil
}
func (w *Wallet)SetUtxos(utxos []Utxo){
	w.Utxos = utxos
}

// AddAddress generates a new payment address and adds it to the wallet.
func (w *Wallet) AddAddress() Address {
	index := uint32(len(w.Skeys))
	newKey := crypto.DeriveSigningKey(w.rootKey, index)
	w.Skeys = append(w.Skeys, newKey)
	return newEnterpriseAddress(newKey.ExtendedVerificationKey(), w.network)
}
func (w *Wallet) AddressIndex(idx int) (Address,error) {
	if idx >=100000{
		return "", errors.New("一个账户不能生产超过10w个地址")
	}
	index := uint32(idx)
	newKey := crypto.DeriveSigningKey(w.rootKey, index)
	//w.Skeys = append(w.Skeys, newKey)
	return newEnterpriseAddress(newKey.ExtendedVerificationKey(), w.network),nil
}
func (w *Wallet) GenAddress(idx int) (addr Address,pri string,err error) {
	if idx >=100000{
		return "","", errors.New("一个账户不能生产超过10w个地址")
	}
	index := uint32(idx)
	newKey := crypto.DeriveSigningKey(w.rootKey, index)
	pri = hex.EncodeToString(newKey)
	//w.Skeys = append(w.Skeys, newKey)
	return newEnterpriseAddress(newKey.ExtendedVerificationKey(), w.network),pri,nil
}
// Addresses returns all wallet's addresss.
func (w *Wallet) Addresses() []Address {
	addresses := make([]Address, len(w.Skeys))
	for i, key := range w.Skeys {
		addresses[i] = newEnterpriseAddress(key.ExtendedVerificationKey(), w.network)
	}
	return addresses
}

func newWalletID() string {
	id, _ := gonanoid.Generate(walleIDAlphabet, 10)
	return "wallet_" + id
}

func NewWallet(name, password string, entropy []byte) *Wallet {
	wallet := &Wallet{Name: name, ID: newWalletID()}
	rootKey := crypto.NewExtendedSigningKey(entropy, password)
	purposeKey := crypto.DeriveSigningKey(rootKey, purposeIndex)
	coinKey := crypto.DeriveSigningKey(purposeKey, coinTypeIndex)
	accountKey := crypto.DeriveSigningKey(coinKey, accountIndex)
	chainKey := crypto.DeriveSigningKey(accountKey, externalChainIndex)
	addr0Key := crypto.DeriveSigningKey(chainKey, 0)
	wallet.rootKey = chainKey
	wallet.Skeys = []crypto.ExtendedSigningKey{addr0Key}
	return wallet
}

type walletDump struct {
	ID      string
	Name    string
	Keys    []crypto.ExtendedSigningKey
	RootKey crypto.ExtendedSigningKey
}

func (w *Wallet) marshal() ([]byte, error) {
	wd := &walletDump{
		ID:      w.ID,
		Name:    w.Name,
		Keys:    w.Skeys,
		RootKey: w.rootKey,
	}
	bytes, err := json.Marshal(wd)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (w *Wallet) unmarshal(bytes []byte) error {
	wd := &walletDump{}
	err := json.Unmarshal(bytes, wd)
	if err != nil {
		return err
	}
	w.ID = wd.ID
	w.Name = wd.Name
	w.Skeys = wd.Keys
	w.rootKey = wd.RootKey
	return nil
}

func ParseUint64(s string) (uint64, error) {
	parsed, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

var NewEntropy = func(bitSize int) []byte {
	entropy, _ := bip39.NewEntropy(bitSize)
	return entropy
}
