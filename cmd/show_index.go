package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var ShowIndexCmd = &cobra.Command{
	Use:   "show-index [table]",
	Short: "Show indexes for a table",
	Long:  `Display all indexes defined on a table.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		tableName := args[0]
		// Validate table name to prevent SQL injection
		if !isValidIdentifier(tableName) {
			return fmt.Errorf("invalid table name: contains potentially dangerous characters")
		}
		rows, err := db.GetDB().Query(fmt.Sprintf("SHOW INDEX FROM `%s`", tableName))
		if err != nil {
			return fmt.Errorf("failed to show indexes: %w", err)
		}
		defer rows.Close()

		printResult(rows)
		return nil
	},
}
