package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var DisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from MySQL server",
	Long:  `Close the current MySQL connection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !db.IsConnected() {
			return fmt.Errorf("not connected to any MySQL server")
		}

		err := db.Disconnect()
		if err != nil {
			return fmt.Errorf("failed to disconnect: %w", err)
		}

		fmt.Println("Successfully disconnected from MySQL")
		return nil
	},
}
