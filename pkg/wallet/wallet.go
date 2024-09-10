package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"math/big"

	common "github.com/blockmandu/pkg/commons"
)

const (
	version    = byte(0x00)
	walletFile = "resources/wallet.dat"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	pubKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	return &Wallet{PrivateKey: *privateKey, PublicKey: pubKey}, nil
}

func (w Wallet) GetAddress() []byte {
	pubHashKey := common.HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubHashKey...)
	checksum := common.Checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := common.Base58Encode(fullPayload)

	return address
}

type _pkey struct {
	D, X, Y *big.Int
}

func (w *Wallet) GobEncode() ([]byte, error) {
	privKey := &_pkey{D: w.PrivateKey.D, X: w.PrivateKey.PublicKey.X, Y: w.PrivateKey.PublicKey.Y}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(privKey)
	if err != nil {
		return nil, err
	}

	_, err = buffer.Write(w.PublicKey)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (w *Wallet) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	var privKey _pkey

	dec := gob.NewDecoder(buf)
	err := dec.Decode(&privKey)
	if err != nil {
		return err
	}

	w.PrivateKey = ecdsa.PrivateKey{D: privKey.D, PublicKey: ecdsa.PublicKey{X: privKey.X, Y: privKey.Y, Curve: elliptic.P256()}}
	w.PublicKey = buf.Bytes()
	return nil
}
