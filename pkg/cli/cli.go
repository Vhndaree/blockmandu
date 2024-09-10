package cli

import (
	"github.com/spf13/cobra"
)

func Run() {
	cmd := &cobra.Command{
		Use:   "blockmandu",
		Short: "The blockmandu is a cli tool for entrypoint of the blockchain.",
	}

	cmd.AddCommand(
		createBlockchainCmd(),
		getBalanceCmd(),
		printChainCmd(),
		sendCmd(),
		createWalletCmd(),
	)

	cobra.CheckErr(cmd.Execute())
}
