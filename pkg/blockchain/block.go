package blockchain

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/blockmandu/pkg/transaction"
)

type Block struct {
	Transactions  []*transaction.Transaction
	PrevBlockHash []byte
	Hash          []byte
	Timestamp     int64
	Nonce         int
}

func NewBlock(txs []*transaction.Transaction, prevBlockHash []byte) (*Block, error) {
	block := &Block{Timestamp: time.Now().Unix(), Transactions: txs, PrevBlockHash: prevBlockHash, Hash: []byte{}, Nonce: 0}
	pow := NewProofOfWork(block)
	nonce, hash, err := pow.Run()
	if err != nil {
		return nil, err
	}
	block.Nonce, block.Hash = nonce, hash[:]

	return block, nil
}

func NewGenesisBlock(coinbase *transaction.Transaction) (*Block, error) {
	return NewBlock([]*transaction.Transaction{coinbase}, []byte{})
}

func (b *Block) HashTransaction() ([]byte, error) {
	var txs [][]byte

	for _, tx := range b.Transactions {
		serialized, err := tx.Serialize()
		if err != nil {
			return nil, err
		}

		txs = append(txs, serialized)
	}

	return NewMerkleTree(txs).RootNode.Data, nil
}

func (b *Block) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func DeserializeBlock(b []byte) (*Block, error) {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(b))
	err := decoder.Decode(&block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}
