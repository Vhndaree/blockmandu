package blockchain

import (
	"bytes"

	common "github.com/blockmandu/pkg/commons"
)

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	return bytes.Equal(common.HashPubKey(in.PubKey), pubKeyHash)
}
