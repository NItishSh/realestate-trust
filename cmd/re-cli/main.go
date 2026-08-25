package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "re-cli",
	Short: "Real Estate Trust Support CLI",
	Long:  `re-cli is a support and RCA tool for the Real Estate Trust & Escrow Platform.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(os.Args) == 1 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
