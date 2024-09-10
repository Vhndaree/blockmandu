package cli

import (
	"log"
	"os"

	"github.com/blockmandu/pkg/blockchain"
	"github.com/spf13/cobra"
)

func createBlockchainCmd() *cobra.Command {
	var address string
	cmd := &cobra.Command{
		Use:   "createblockchain",
		Short: "Create a blockchain with genesis block",
		Run: func(cmd *cobra.Command, args []string) {
			if address == "" {
				cmd.Usage()
				os.Exit(1)
			}

			bc, err := blockchain.CreateBlockchain(address)
			if err != nil {
				log.Panic(err)
			}

			defer bc.DB.Close()
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "", "The address to send genesis block reward to")

	return cmd
}
