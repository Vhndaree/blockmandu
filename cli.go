package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type CLI struct {
	blockchain *Blockchain
}

func NewCli(bc *Blockchain) *CLI {
	return &CLI{blockchain: bc}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add-block", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print-chain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "add-block":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "print-chain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}

		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  add-block -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println("  print-chain - print all the blocks of the blockchain")
}

func (cli *CLI) addBlock(data string) {
	cli.blockchain.AddBlock(data)
}

func (cli *CLI) printChain() {
	bci := cli.blockchain.Iterator()

	for {
		block := bci.Next()
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
