package transaction

import (
	"bytes"
	"encoding/gob"

	common "github.com/blockmandu/pkg/commons"
)

type TXOutput struct {
	PubKeyHash []byte
	Value      int
}

func NewTXOutput(value int, address string) *TXOutput {
	txout := &TXOutput{Value: value, PubKeyHash: nil}
	txout.Lock([]byte(address))

	return txout
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := common.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

type TXOutputs struct {
	Outputs []TXOutput
}

func (o TXOutputs) Serialize() ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(o)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func DeserializeOutputs(data []byte) (TXOutputs, error) {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		return TXOutputs{}, err
	}

	return outputs, nil
}
