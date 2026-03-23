package cmd

import (
	"fmt"
	"os"

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
		_, err := db.GetDB().Exec(fmt.Sprintf("USE %s", escapeIdentifier(dbName)))
		if err != nil {
			return fmt.Errorf("failed to use database: %w", err)
		}

		db.SetCurrentDatabase(dbName)
		if err := db.SaveConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
		}
		fmt.Printf("Switched to database: %s\n", dbName)
		return nil
	},
}
