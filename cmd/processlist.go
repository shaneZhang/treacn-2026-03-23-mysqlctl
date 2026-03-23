package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var ShowProcesslistCmd = &cobra.Command{
	Use:   "processlist",
	Short: "Show running MySQL processes",
	Long:  `Display all currently running processes/queries in the MySQL server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		rows, err := db.GetDB().Query("SHOW FULL PROCESSLIST")
		if err != nil {
			return fmt.Errorf("failed to get processlist: %w", err)
		}
		defer rows.Close()

		printResult(rows)
		return nil
	},
}
