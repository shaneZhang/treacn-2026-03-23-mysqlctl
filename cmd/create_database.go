package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var CreateDatabaseCmd = &cobra.Command{
	Use:   "create-database [name]",
	Short: "Create a new database",
	Long:  `Create a new database on the MySQL server.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		dbName := args[0]
		_, err := db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE %s", escapeIdentifier(dbName)))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}

		fmt.Printf("Database '%s' created successfully\n", dbName)
		return nil
	},
}
