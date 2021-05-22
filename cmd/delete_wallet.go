package cmd

import (
	"github.com/echovl/cardano-wallet/logger"
	"github.com/echovl/cardano-wallet/wallet"
	"github.com/spf13/cobra"
)

// deleteWalletCmd represents the deleteWallet command
var deleteWalletCmd = &cobra.Command{
	Use:     "delete-wallet [wallet-id]",
	Short:   "Delete a wallet",
	Aliases: []string{"delw"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		walletId := wallet.WalletID(args[0])
		logger.Infof("Wallet removed", "wallet", walletId)
		return wallet.DeleteWallet(walletId, DefaultDb)
	},
}

func init() {
	rootCmd.AddCommand(deleteWalletCmd)
}
