package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var ShowTablesCmd = &cobra.Command{
	Use:   "tables",
	Short: "List all tables in current database",
	Long:  `Show all tables in the currently selected database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		rows, err := db.GetDB().Query("SHOW TABLES")
		if err != nil {
			return fmt.Errorf("failed to get tables: %w", err)
		}
		defer rows.Close()

		fmt.Println("Tables:")
		fmt.Println(strings.Repeat("-", 40))
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return fmt.Errorf("failed to scan table name: %w", err)
			}
			fmt.Println(tableName)
		}
		return nil
	},
}
