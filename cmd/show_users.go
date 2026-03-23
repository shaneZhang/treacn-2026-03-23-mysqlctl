package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var ShowUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Show all MySQL users",
	Long:  `Display all user accounts in the MySQL server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		rows, err := db.GetDB().Query("SELECT User, Host FROM mysql.user")
		if err != nil {
			return fmt.Errorf("failed to get users: %w", err)
		}
		defer rows.Close()

		fmt.Println("Users:")
		fmt.Println(strings.Repeat("-", 40))
		for rows.Next() {
			var user, host string
			if err := rows.Scan(&user, &host); err != nil {
				return fmt.Errorf("failed to scan user: %w", err)
			}
			fmt.Printf("%s@%s\n", user, host)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating users: %w", err)
		}
		return nil
	},
}
