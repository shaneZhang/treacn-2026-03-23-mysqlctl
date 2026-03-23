package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var UseDatabaseCmd = &cobra.Command{
	Use:   "use [database]",
	Short: "Switch to a database",
	Long:  `Select a database to use.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		dbName := args[0]
		err := db.UpdateDatabase(dbName)
		if err != nil {
			return err
		}

		fmt.Printf("Switched to database: %s\n", dbName)
		return nil
	},
}
