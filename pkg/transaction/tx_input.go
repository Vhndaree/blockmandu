package transaction

import (
	"bytes"

	common "github.com/blockmandu/pkg/commons"
)

type TXInput struct {
	Txid      []byte
	Signature []byte
	PubKey    []byte
	Vout      int
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	return bytes.Equal(common.HashPubKey(in.PubKey), pubKeyHash)
}
