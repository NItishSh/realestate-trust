package main

import (
	"fmt"
	"log"

	"github.com/realestate-trust/monorepo/internal/db"
	"github.com/spf13/cobra"
)

var ledgerCmd = &cobra.Command{
	Use:   "ledger",
	Short: "Ledger operations",
}

var ledgerVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify the cryptographic chain for the entire ledger",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Verifying entire ledger chain...")

		database, err := db.Connect()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer func() { _ = database.Close() }()

		if database.SQL == nil {
			log.Fatalf("No DATABASE_URL configured")
		}

		repo := db.NewSQLLedgerRepository(database.SQL)
		logs, err := repo.ListLogs()
		if err != nil {
			log.Fatalf("Failed to fetch logs: %v", err)
		}

		if len(logs) == 0 {
			fmt.Println("Ledger is empty.")
			return
		}

		var prevHash string
		for _, entry := range logs {
			if entry.PreviousHash != prevHash {
				log.Fatalf("Chain broken at index %d! Expected PrevHash %s, got %s", entry.Index, prevHash, entry.PreviousHash)
			}
			calcHash := entry.CalculateHash()
			if entry.Hash != calcHash {
				log.Fatalf("Data corruption at index %d! Stored Hash %s does not match calculated %s", entry.Index, entry.Hash, calcHash)
			}
			prevHash = entry.Hash
		}

		fmt.Printf("Verification: PASSED (%d entries verified)\n", len(logs))
	},
}

func init() {
	ledgerCmd.AddCommand(ledgerVerifyCmd)
	rootCmd.AddCommand(ledgerCmd)
}
