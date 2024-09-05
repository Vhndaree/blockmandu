package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
)

type TXInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

type TXOutput struct {
	Value        int
	ScriptPubKey string
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

const subsidy = 10

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = "Reward to '%s'" + to
	}

	txin := TXInput{Txid: []byte{}, Vout: -1, ScriptSig: data}
	txout := TXOutput{Value: subsidy, ScriptPubKey: to}

	tx := Transaction{ID: nil, Vin: []TXInput{txin}, Vout: []TXOutput{txout}}
	tx.SetID()

	return &tx
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func NewUTXOTransaction(from, to string, amount int, blockchain *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := blockchain.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{Txid: txID, Vout: out, ScriptSig: from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOutput{Value: amount, ScriptPubKey: to})
	if acc > amount {
		outputs = append(outputs, TXOutput{Value: acc - amount, ScriptPubKey: from})
	}

	tx := Transaction{ID: nil, Vin: inputs, Vout: outputs}
	tx.SetID()
	return &tx
}
