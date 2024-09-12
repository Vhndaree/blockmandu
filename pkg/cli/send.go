package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/blockmandu/pkg/blockchain"
	common "github.com/blockmandu/pkg/commons"
	"github.com/spf13/cobra"
)

func sendCmd() *cobra.Command {
	var to, from string
	var amount int
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send blockmandu to given address",
		Run: func(cmd *cobra.Command, args []string) {
			if amount <= 0 {
				cmd.Usage()
				os.Exit(1)
			}

			send(from, to, amount)
		},
	}

	cmd.Flags().StringVarP(&to, "to", "", "", "Destination wallet address")
	cmd.Flags().StringVarP(&from, "from", "", "", "Source wallet address")
	cmd.Flags().IntVarP(&amount, "amount", "a", 0, "Amount to be sent")

	return cmd
}

func send(from, to string, amount int) {
	if !common.ValidateAddress(from) {
		log.Panic("Err: Sender address is not valid")
	}

	if !common.ValidateAddress(to) {
		log.Panic("Err: Recipient address is not valid")
	}

	if from == to {
		log.Panic("Err: Sender and Recipient address cannot be the same")
	}

	bc, err := blockchain.NewBlockchain(from)
	if err != nil {
		log.Panic(err)
	}
	defer bc.DB.Close()

	tx, err := blockchain.NewUTXOTransaction(from, to, amount, bc)
	if err != nil {
		log.Panic(err)
	}

	err = bc.MineBlock([]*blockchain.Transaction{tx})
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Success!")
}
