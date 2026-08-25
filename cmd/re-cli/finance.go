package main

import (
	"fmt"
	"log"

	"github.com/realestate-trust/monorepo/internal/db"
	"github.com/spf13/cobra"
)

var financeCmd = &cobra.Command{
	Use:   "finance",
	Short: "Financing operations",
}

var financeCheckCmd = &cobra.Command{
	Use:   "check [loan_id]",
	Short: "Check a loan application status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		loanID := args[0]
		fmt.Printf("Checking finance application: %s\n", loanID)

		database, err := db.Connect()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer func() { _ = database.Close() }()

		if database.SQL == nil {
			log.Fatalf("No DATABASE_URL configured")
		}

		repo := db.NewSQLFinancingRepository(database.SQL)
		loan, err := repo.GetLoan(loanID)
		if err != nil {
			log.Fatalf("Failed to get loan: %v", err)
		}

		fmt.Printf("ID: %s\n", loan.ID)
		fmt.Printf("Transaction ID: %s\n", loan.TransactionID)
		fmt.Printf("User ID: %s\n", loan.UserID)
		fmt.Printf("Requested Amount: %.2f\n", loan.RequestedAmount)
		if loan.ApprovedAmount != nil {
			fmt.Printf("Approved Amount: %.2f\n", *loan.ApprovedAmount)
		} else {
			fmt.Printf("Approved Amount: N/A\n")
		}
		fmt.Printf("Status: %s\n", loan.Status)
	},
}

func init() {
	financeCmd.AddCommand(financeCheckCmd)
	rootCmd.AddCommand(financeCmd)
}
