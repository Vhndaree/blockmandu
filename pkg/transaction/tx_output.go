package transaction

import (
	"bytes"

	common "github.com/blockmandu/pkg/commons"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
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
