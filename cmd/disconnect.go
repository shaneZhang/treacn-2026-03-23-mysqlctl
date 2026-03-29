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
		// 清除配置文件
		if err := db.ClearConfigFile(); err != nil {
			return fmt.Errorf("failed to clear config: %w", err)
		}

		fmt.Println("Successfully disconnected from MySQL")
		return nil
	},
}
