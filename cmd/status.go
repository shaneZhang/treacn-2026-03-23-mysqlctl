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
		// 尝试从配置文件加载状态
		cfg, err := db.LoadConfigFromFile()
		if err != nil || cfg == nil {
			fmt.Println("Connection Status: Not connected")
			return nil
		}

		fmt.Println("Connection Status: Connected (from saved config)")
		fmt.Printf("Server: %s:%d\n", cfg.Host, cfg.Port)
		fmt.Printf("User: %s\n", cfg.User)
		fmt.Printf("Database: %s\n", cfg.Database)
		return nil
	},
}
