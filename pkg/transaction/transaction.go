package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math/big"
)

const subsidy = 10

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func NewCoinbaseTX(to, data string) (*Transaction, error) {
	if data == "" {
		data = "Reward to '%s'" + to
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	txid, err := tx.Hash()
	if err != nil {
		return nil, err
	}

	tx.ID = txid
	return &tx, nil
}

func (tx *Transaction) Hash() ([]byte, error) {
	txCopy := *tx
	txCopy.ID = []byte{}

	serialized, err := txCopy.Serialize()
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(serialized)

	return hash[:], nil
}

func (tx *Transaction) Serialize() ([]byte, error) {
	var encoded bytes.Buffer

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	if err != nil {
		return nil, err
	}

	return encoded.Bytes(), nil
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txid, err := txCopy.Hash()
		if err != nil {
			return err
		}

		txCopy.ID = txid
		txCopy.Vin[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			return err
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}

	return nil
}

func (tx Transaction) Verify(prevTxs map[string]Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	for _, vin := range tx.Vin {
		if prevTxs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTxs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txid, err := txCopy.Hash()
		if err != nil {
			return false, err
		}

		txCopy.ID = txid
		txCopy.Vin[inID].PubKey = nil

		r, s := big.Int{}, big.Int{}
		siglen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(siglen / 2)])
		s.SetBytes(vin.Signature[(siglen / 2):])

		x, y := big.Int{}, big.Int{}
		keylen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keylen / 2)])
		y.SetBytes(vin.PubKey[(keylen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false, nil
		}
	}

	return true, nil
}

func (tx Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{Txid: vin.Txid, Vout: vin.Vout, Signature: nil, PubKey: nil})
	}

	return Transaction{ID: tx.ID, Vin: inputs, Vout: tx.Vout}
}
