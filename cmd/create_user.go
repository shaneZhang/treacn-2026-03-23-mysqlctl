package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var (
	createUserHost     string
	createUserPassword string
)

var CreateUserCmd = &cobra.Command{
	Use:   "create-user [username]",
	Short: "Create a new MySQL user",
	Long:  `Create a new user account.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		username := args[0]
		host := createUserHost
		if host == "" {
			host = "%"
		}

		var sqlQuery string
		if createUserPassword != "" {
			sqlQuery = fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'",
				username, host, createUserPassword)
		} else {
			sqlQuery = fmt.Sprintf("CREATE USER '%s'@'%s'",
				username, host)
		}

		_, err := db.GetDB().Exec(sqlQuery)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		fmt.Printf("User '%s'@'%s' created successfully\n", username, host)
		return nil
	},
}

func init() {
	CreateUserCmd.Flags().StringVar(&createUserHost, "host", "%", "User host (default: %)")
	CreateUserCmd.Flags().StringVarP(&createUserPassword, "password", "p", "", "User password")
}
