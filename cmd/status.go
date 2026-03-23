package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show connection status",
	Long:  `Display the current MySQL connection status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		connected, dbName := db.GetStatus()
		if !connected {
			_ = db.LoadAndConnect()
			connected, dbName = db.GetStatus()
		}
		if connected {
			cfg := db.GetConfig()
			fmt.Println("Connection Status: Connected")
			fmt.Printf("Server: %s:%d\n", cfg.Host, cfg.Port)
			fmt.Printf("User: %s\n", cfg.User)
			fmt.Printf("Database: %s\n", dbName)
		} else {
			fmt.Println("Connection Status: Not connected")
		}
		return nil
	},
}
