package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var QueryCmd = &cobra.Command{
	Use:   "query [sql]",
	Short: "Execute a SQL query",
	Long:  `Execute an arbitrary SQL query and display results.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		sqlQuery := args[0]
		rows, err := db.GetDB().Query(sqlQuery)
		if err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()

		printResult(rows)
		return nil
	},
}
