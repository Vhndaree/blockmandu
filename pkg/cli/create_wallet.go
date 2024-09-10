package cli

import (
	"fmt"
	"log"

	"github.com/blockmandu/pkg/wallet"
	"github.com/spf13/cobra"
)

func createWalletCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createwallet",
		Short: "Create a blockchain with genesis block",
		Run: func(cmd *cobra.Command, args []string) {
			wallets, err := wallet.NewWallets()
			if err != nil {
				log.Panic(err)
			}

			address, err := wallets.CreateWallet()
			if err != nil {
				log.Panic(err)
			}

			wallets.SaveToFile()
			fmt.Printf("Your new address: %s\n", address)
		},
	}

	return cmd
}
