package main

type Blockchain struct {
	blocks []*Block
}

func NewBlockChain() *Blockchain {
	genesisBlock := NewBlock("Genesis Block!", []byte{})
	return &Blockchain{blocks: []*Block{genesisBlock}}
}

func (bc *Blockchain) AddBlocks(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}
