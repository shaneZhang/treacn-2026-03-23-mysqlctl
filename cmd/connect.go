package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var ConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a MySQL server",
	Long:  `Establish a connection to a MySQL database server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("password")
		database, _ := cmd.Flags().GetString("database")

		if host == "" {
			return fmt.Errorf("host is required")
		}
		if user == "" {
			return fmt.Errorf("username is required")
		}

		err := db.Connect(host, port, user, password, database)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		fmt.Printf("Successfully connected to MySQL at %s:%d\n", host, port)
		if database != "" {
			fmt.Printf("Using database: %s\n", database)
		}
		return nil
	},
}
