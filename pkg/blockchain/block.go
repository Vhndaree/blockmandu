package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{Timestamp: time.Now().Unix(), Transactions: txs, PrevBlockHash: prevBlockHash, Hash: []byte{}, Nonce: 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Nonce, block.Hash = nonce, hash[:]

	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		panic("unable to serialize block" + err.Error())
	}

	return result.Bytes()
}

func Deserialize(b []byte) (*Block, error) {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(b))
	err := decoder.Decode(&block)
	if err != nil {
		return nil, err
	}

	return &block, nil
}
