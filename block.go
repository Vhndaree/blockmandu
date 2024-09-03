package main

import (
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{Timestamp: time.Now().Unix(), Data: []byte(data), PrevBlockHash: prevBlockHash, Hash: []byte{}, Nonce: 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Nonce, block.Hash = nonce, hash[:]

	return block
}
