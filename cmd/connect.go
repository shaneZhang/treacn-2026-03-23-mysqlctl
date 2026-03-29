package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var connectHost string
var connectPort int
var connectUser string
var connectPassword string
var connectDatabase string

var ConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a MySQL server",
	Long:  `Establish a connection to a MySQL database server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if connectHost == "" {
			return fmt.Errorf("host is required")
		}
		if connectUser == "" {
			return fmt.Errorf("username is required")
		}
		if connectPort == 0 {
			connectPort = 3306
		}

		err := db.Connect(connectHost, connectPort, connectUser, connectPassword, connectDatabase)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		// 保存配置到文件以便其他命令使用
		cfg := &db.Config{
			Host:     connectHost,
			Port:     connectPort,
			User:     connectUser,
			Password: connectPassword,
			Database: connectDatabase,
		}
		if err := db.SaveConfigToFile(cfg); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to save config: %v\n", err)
		}

		fmt.Printf("Successfully connected to MySQL at %s:%d\n", connectHost, connectPort)
		if connectDatabase != "" {
			fmt.Printf("Using database: %s\n", connectDatabase)
		}
		return nil
	},
}

func init() {
	ConnectCmd.Flags().StringVar(&connectHost, "host", "localhost", "MySQL server host")
	ConnectCmd.Flags().IntVarP(&connectPort, "port", "P", 3306, "MySQL server port")
	ConnectCmd.Flags().StringVarP(&connectUser, "user", "u", "root", "MySQL username")
	ConnectCmd.Flags().StringVarP(&connectPassword, "password", "p", "", "MySQL password")
	ConnectCmd.Flags().StringVarP(&connectDatabase, "database", "d", "", "Database name to use")
}
