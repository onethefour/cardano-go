package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/echovl/cardano-go"
)

type CardanoCli struct{}

type CardanoCliTip struct {
	Epoch uint64
	Hash  string
	Slot  uint64
	Block uint64
	Era   string
}
type CardanoCliTx struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CborHex     string `json:"cborHex"`
}

func (cli *CardanoCli) QueryUtxos(address cardano.Address) ([]cardano.Utxo, error) {
	out, err := runCommand("cardano-cli", "query", "utxo", "--address", string(address), "--testnet-magic", "1097911063")
	if err != nil {
		return nil, err
	}

	counter := 0
	utxos := []cardano.Utxo{}
	for {
		line, err := out.ReadString(byte('\n'))
		if err != nil {
			break
		}
		if counter >= 2 {
			args := strings.Fields(line)
			if len(args) < 4 {
				return nil, fmt.Errorf("malformed cli response")
			}

			txId := cardano.TxId(args[0])
			index, err := cardano.ParseUint64(args[1])
			if err != nil {
				return nil, err
			}
			amount, err := cardano.ParseUint64(args[2])
			if err != nil {
				return nil, err
			}

			utxos = append(utxos, cardano.Utxo{
				TxId:    txId,
				Index:   index,
				Amount:  amount,
				Address: address,
			})
		}
		counter++
	}

	return utxos, nil
}

func (cli *CardanoCli) QueryTip() (cardano.NodeTip, error) {
	out, err := runCommand("cardano-cli", "query", "tip", "--testnet-magic", "1097911063")
	if err != nil {
		return cardano.NodeTip{}, err
	}

	cliTip := &CardanoCliTip{}
	err = json.Unmarshal(out.Bytes(), cliTip)
	if err != nil {
		return cardano.NodeTip{}, err
	}

	return cardano.NodeTip{
		Epoch: cliTip.Epoch,
		Block: cliTip.Block,
		Slot:  cliTip.Slot,
	}, nil
}

func (cli *CardanoCli) SubmitTx(tx cardano.Transaction) error {
	const txFileName = "txsigned.temp"
	txPayload := CardanoCliTx{
		Type:        "Tx MaryEra",
		Description: "",
		CborHex:     tx.CborHex(),
	}

	txPayloadJson, err := json.Marshal(txPayload)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(txFileName, txPayloadJson, 777)
	if err != nil {
		return err
	}

	out, err := runCommand("cardano-cli", "transaction", "submit", "--tx-file", txFileName, "--testnet-magic", "1097911063")
	fmt.Print(out.String())

	err = os.Remove(txFileName)

	return err
}

func runCommand(cmd string, arg ...string) (*bytes.Buffer, error) {
	out := &bytes.Buffer{}
	command := exec.Command(cmd, arg...)
	command.Stdout = out
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return nil, err
	}

	return out, nil
}
