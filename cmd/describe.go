package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var DescribeTableCmd = &cobra.Command{
	Use:   "describe [table]",
	Short: "Show table structure",
	Long:  `Display the structure of a table (columns, types, keys, etc.).`,
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
		rows, err := db.GetDB().Query(fmt.Sprintf("DESCRIBE `%s`", tableName))
		if err != nil {
			return fmt.Errorf("failed to describe table: %w", err)
		}
		defer rows.Close()

		printResult(rows)
		return nil
	},
}
