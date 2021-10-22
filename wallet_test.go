package cardano

import (
	"encoding/hex"
	"github.com/onethefour/cardano-go/crypto"
	"io/ioutil"
	"testing"

	"github.com/echovl/bech32"
	"github.com/tyler-smith/go-bip39"
)
func TestNewWallet(t *testing.T) {
	pri := NewEntropy(entropySizeInBits)
	mnemonic := crypto.NewMnemonic(pri)
	//pri := make([]byte,entropySizeInBits)
	//n,err := rand.Read(pri)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//if n != entropySizeInBits{
	//	t.Fatal()
	//}

	wallet := NewWallet("goapi","",pri)
	wallet.SetNetwork(Mainnet)
	wallet.AddAddress()
	t.Log(wallet.AddressIndex(1))
	t.Log(len(wallet.Skeys),wallet.Addresses())
	t.Log(mnemonic)
}
func TestWallet(t *testing.T){
	mnemonic :=  "lunch crisp short cliff wear ask lend below supply grab base shrug portion coconut ketchup"
	pri,err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatal(err.Error())
	}
	wallet := NewWallet("goapi","",pri)
	wallet.SetNetwork(Mainnet)
	wallet.AddAddress()
	t.Log(wallet.GenAddress(0))
	t.Log(wallet.GenAddress(1))
	wallet.SetKey("90798195e93127dc4d7b3a462eb52fa1557219647c1d073e8c52af133d82434cea45c431aef729112679a3d9958dea95d191f25d22966ea12fb0316a4752fbb85b67ec845a6f91bfd0e6736a23a9f6f0a8c661c65182b2688cfe12121ad933c0")
	wallet.SetKey("e0c5143cf526093c871678015d09b83409a224c30f8d594f5972d8c84082434cf1a16668e6414714936c96578110b6c115c6fbd94fed71f9765bba1fbbcb1eed48b5f0eb6755816a725c557c3c8dafbd315e5d03fc505a20fad0d5c1d33cd737")
	t.Log(len(wallet.Skeys),wallet.Addresses())
	//wallet_test.go:41: 2 [addr1vx69t76fpgm6p5jf33295sculs7d7ca4z0cse9chucg2ufgmakpa0 addr1vyjhlavlwgmxgukze8l32syws4ft6kqtkl0yca747dek8pcvhhv08]

	t.Log(mnemonic)
	toAddr,err := Bech32ToAddress("addr1vyjhlavlwgmxgukze8l32syws4ft6kqtkl0yca747dek8pcvhhv08")
	if err != nil {
		t.Fatal(err.Error())
	}
	changeAddr,err := Bech32ToAddress("addr1vx69t76fpgm6p5jf33295sculs7d7ca4z0cse9chucg2ufgmakpa0")
	if err != nil {
		t.Fatal(err.Error())
	}
	utxos := make([]Utxo,0)
	utxo,err := NewUtxo("64c1d5f99f02326023d881f9b1b9484bdead18a2ab2edf231855ebe229acc7ab","addr1vx69t76fpgm6p5jf33295sculs7d7ca4z0cse9chucg2ufgmakpa0",0,9000000)
	if err != nil{
		t.Fatal(err.Error())
	}
	utxos = append(utxos,utxo)
	wallet.Utxos = utxos
	wallet.Tip = NodeTip{
		Epoch:0,
		Block:6385709,
		Slot: 42974969,
	}
	tx,err := wallet.Transfer(toAddr,1000000,changeAddr)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(hex.EncodeToString(tx.Bytes()))
	ioutil.WriteFile("/Users/one/xutonghua/xth_signserver/adasign/signedtx.text", tx.Bytes(), 0666)
//83a4008182582064c1d5f99f02326023d881f9b1b9484bdead18a2ab2edf231855ebe229acc7ab00018282581d61257ff59f72366472c2c9ff15408e8552bd580bb7de4c77d5f37363871a000f424082581d61b455fb490a37a0d2498c545a431cfc3cdf63b513f10c9717e610ae251a00778a4f021a000287b1031a028fc3a9a1008182582044b64200797bb67176084116826a9462226f0b3498e0c15446370db28372486458408c7de8a9e34ef80ffb5e083e0c5cb2e0b66c46422ca33f683286a3a7892bacedcdd128c05f12957b9b79f773e4d3fc4efc058cbb57b24a0f03c6a729e50a6309f6
}

//wallet_test.go:26: 2 [addr1vx69t76fpgm6p5jf33295sculs7d7ca4z0cse9chucg2ufgmakpa0 addr1vyjhlavlwgmxgukze8l32syws4ft6kqtkl0yca747dek8pcvhhv08]

func TestGenerateAddress(t *testing.T) {
	for _, testVector := range testVectors {
		client := NewClient(WithDB(&MockDB{}))
		defer client.Close()

		NewEntropy = func(bitSize int) []byte {
			entropy, err := bip39.EntropyFromMnemonic(testVector.mnemonic)
			if err != nil {
				t.Error(err)
			}
			return entropy
		}

		w, _, err := client.CreateWallet("test", "")
		if err != nil {
			t.Error(err)
		}
		w.SetNetwork(Testnet)

		paymentAddr1 := w.AddAddress()

		addrXsk1 := bech32From("addr_xsk", w.Skeys[1])
		addrXvk1 := bech32From("addr_xvk", w.Skeys[1].ExtendedVerificationKey())

		if addrXsk1 != testVector.addrXsk1 {
			t.Errorf("invalid addrXsk1 :\ngot: %v\nwant: %v", addrXsk1, testVector.addrXsk1)
		}

		if addrXvk1 != testVector.addrXvk1 {
			t.Errorf("invalid addrXvk1 :\ngot: %v\nwant: %v", addrXvk1, testVector.addrXvk1)
		}

		if paymentAddr1 != testVector.paymentAddr1 {
			t.Errorf("invalid paymentAddr1:\ngot: %v\nwant: %v", paymentAddr1, testVector.paymentAddr1)
		}
	}
}

type MockNode struct {
	utxos []Utxo
}

func (prov *MockNode) QueryUtxos(addr Address) ([]Utxo, error) {
	return prov.utxos, nil
}

func (prov *MockNode) QueryTip() (NodeTip, error) {
	return NodeTip{}, nil
}

func (prov *MockNode) SubmitTx(tx transaction) error {
	return nil
}

func TestWalletBalance(t *testing.T) {
	client := NewClient(WithDB(&MockDB{}))
	client.node = &MockNode{utxos: []Utxo{{Amount: 100}, {Amount: 33}}}
	w, _, err := client.CreateWallet("test", "")
	if err != nil {
		t.Error(err)
	}

	got, err := w.Balance()
	if err != nil {
		t.Error(err)
	}
	want := uint64(133)

	if got != want {
		t.Errorf("invalid balance :\ngot: %v\nwant: %v", got, want)
	}
}

func bech32From(hrp string, bytes []byte) string {
	enc, _ := bech32.EncodeFromBase256(hrp, bytes)
	return enc
}
