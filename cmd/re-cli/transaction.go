package main

import (
	"fmt"
	"log"

	"github.com/realestate-trust/monorepo/internal/db"
	"github.com/spf13/cobra"
)

var transactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Transaction operations",
}

var transactionInspectCmd = &cobra.Command{
	Use:   "inspect [txn_id]",
	Short: "Inspect a transaction's current state",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		txnID := args[0]
		fmt.Printf("Inspecting transaction: %s\n", txnID)

		database, err := db.Connect()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer func() { _ = database.Close() }()

		if database.SQL == nil {
			log.Fatalf("No DATABASE_URL configured")
		}

		repo := db.NewSQLTransactionRepository(database.SQL)
		txn, err := repo.GetTransaction(txnID)
		if err != nil {
			log.Fatalf("Failed to get transaction: %v", err)
		}

		fmt.Printf("ID: %s\n", txn.ID)
		fmt.Printf("Property ID: %s\n", txn.PropertyID)
		fmt.Printf("Buyer ID: %s\n", txn.BuyerID)
		fmt.Printf("Seller ID: %s\n", txn.SellerID)
		fmt.Printf("Total Amount: %.2f\n", txn.TotalAmount)
		fmt.Printf("Status: %s\n", txn.Status)
		fmt.Printf("Virtual Account: %s\n", txn.VirtualAccountNumber)
	},
}

func init() {
	transactionCmd.AddCommand(transactionInspectCmd)
	rootCmd.AddCommand(transactionCmd)
}
