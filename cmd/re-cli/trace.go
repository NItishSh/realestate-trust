package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace [txn_id]",
	Short: "Trace a transaction across microservices",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		txnID := args[0]
		// TODO: Implement integration with Loki/Tempo API
		fmt.Printf("Fetching traces for transaction: %s\n", txnID)
		fmt.Println("No traces found (stub)")
	},
}

func init() {
	rootCmd.AddCommand(traceCmd)
}
