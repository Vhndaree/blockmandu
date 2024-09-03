package main

import (
	"fmt"
	"strconv"
)

func main() {
	bc := NewBlockChain()

	bc.AddBlocks("Send 1 Blockmandu to Ivan")
	bc.AddBlocks("Send 2 more Blockmandu to Ivan")

	for _, block := range bc.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		fmt.Println()
	}
}
