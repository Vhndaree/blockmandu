package wallet

import (
	"crypto/ecdh"
	"crypto/rand"

	common "github.com/blockmandu/pkg/commons"
)

const (
	version    = byte(0x00)
	walletFile = "resources/wallet.dat"
)

type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

func NewWallet() (*Wallet, error) {
	curve := ecdh.P256()
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pubKey := privateKey.PublicKey().Bytes()

	return &Wallet{PrivateKey: privateKey.Bytes(), PublicKey: pubKey}, nil
}

func (w Wallet) GetAddress() []byte {
	pubHashKey := common.HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubHashKey...)
	checksum := common.Checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := common.Base58Encode(fullPayload)

	return address
}
