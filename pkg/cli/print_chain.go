package cli

import (
	"fmt"
	"log"

	"github.com/blockmandu/pkg/blockchain"
	"github.com/spf13/cobra"
)

func printChainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "printchain",
		Short: "Display complete blockchain",
		Run: func(cmd *cobra.Command, args []string) {
			printChain()
		},
	}
}

func printChain() {
	bc, err := blockchain.NewBlockchain("")
	if err != nil {
		log.Panic(err)
	}

	defer bc.DB.Close()

	bci := bc.Iterator()
	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProofOfWork(block)

		valid, _ := pow.Validate()
		fmt.Printf("PoW: %t\n\n", valid)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
