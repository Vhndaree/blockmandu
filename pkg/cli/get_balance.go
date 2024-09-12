package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/blockmandu/pkg/blockchain"
	common "github.com/blockmandu/pkg/commons"
	"github.com/spf13/cobra"
)

func getBalanceCmd() *cobra.Command {
	var address string
	cmd := &cobra.Command{
		Use:   "getbalance",
		Short: "Get balance of a given address",
		Run: func(cmd *cobra.Command, args []string) {
			if address == "" {
				cmd.Usage()
				os.Exit(1)
			}

			if !common.ValidateAddress(address) {
				log.Panic("ERROR: Address is not valid")
			}

			getBalance(address)
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "", "The address to get balance for")

	return cmd
}

func getBalance(address string) {
	bc, err := blockchain.NewBlockchain(address)
	if err != nil {
		log.Panic(err)
	}
	defer bc.DB.Close()

	balance := 0
	pubKeyHash := common.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := bc.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
