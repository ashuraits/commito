package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update commito to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := selfUpdate(); err != nil {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
