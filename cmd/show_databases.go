package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var ShowDatabasesCmd = &cobra.Command{
	Use:   "databases",
	Short: "List all databases",
	Long:  `Show all databases on the MySQL server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		rows, err := db.GetDB().Query("SHOW DATABASES")
		if err != nil {
			return fmt.Errorf("failed to get databases: %w", err)
		}
		defer rows.Close()

		fmt.Println("Databases:")
		fmt.Println(strings.Repeat("-", 40))
		for rows.Next() {
			var dbName string
			if err := rows.Scan(&dbName); err != nil {
				return fmt.Errorf("failed to scan database name: %w", err)
			}
			fmt.Println(dbName)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating databases: %w", err)
		}
		return nil
	},
}
