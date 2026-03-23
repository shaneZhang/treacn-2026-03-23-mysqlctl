package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var DropDatabaseCmd = &cobra.Command{
	Use:   "drop-database [name]",
	Short: "Drop a database",
	Long:  `Delete a database from the MySQL server.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		dbName := args[0]

		// Check if we're currently using this database, if so clear it
		_, currentDb := db.GetStatus()
		if currentDb == dbName {
			db.UpdateDatabase("")
		}

		_, err := db.GetDB().Exec(fmt.Sprintf("DROP DATABASE `%s`", dbName))
		if err != nil {
			return fmt.Errorf("failed to drop database: %w", err)
		}

		fmt.Printf("Database '%s' dropped successfully\n", dbName)
		return nil
	},
}
