package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var (
	grantUser     string
	grantHost     string
	grantPriv     string
	grantDatabase string
	grantTable    string
)

var GrantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Grant privileges to a user",
	Long:  `Grant privileges to a MySQL user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		if grantUser == "" {
			return fmt.Errorf("username is required (use --user)")
		}
		if grantPriv == "" {
			return fmt.Errorf("privileges are required (use --privileges)")
		}
		// Validate that database is specified when using --table flag
		if grantTable != "" && grantDatabase == "" {
			return fmt.Errorf("--database is required when using --table flag")
		}

		host := grantHost
		if host == "" {
			host = "%"
		}

		// Validate identifiers to prevent SQL injection
		if !isValidIdentifier(grantUser) {
			return fmt.Errorf("invalid username: contains potentially dangerous characters")
		}
		if !isValidHost(host) {
			return fmt.Errorf("invalid host: contains potentially dangerous characters")
		}
		if !isValidPrivilege(grantPriv) {
			return fmt.Errorf("invalid privileges: contains potentially dangerous characters")
		}

		target := grantDatabase
		if grantTable != "" {
			// Validate database and table names
			if !isValidIdentifier(grantDatabase) {
				return fmt.Errorf("invalid database name: contains potentially dangerous characters")
			}
			if !isValidIdentifier(grantTable) {
				return fmt.Errorf("invalid table name: contains potentially dangerous characters")
			}
			target = fmt.Sprintf("`%s`.`%s`", grantDatabase, grantTable)
		} else if grantDatabase != "" && grantDatabase != "*" {
			if !isValidIdentifier(grantDatabase) {
				return fmt.Errorf("invalid database name: contains potentially dangerous characters")
			}
			target = fmt.Sprintf("`%s`.*", grantDatabase)
		} else {
			target = "*.*"
		}

		sqlQuery := fmt.Sprintf("GRANT %s ON %s TO '%s'@'%s'",
			grantPriv, target, grantUser, host)

		_, err := db.GetDB().Exec(sqlQuery)
		if err != nil {
			return fmt.Errorf("failed to grant privileges: %w", err)
		}

		_, err = db.GetDB().Exec("FLUSH PRIVILEGES")
		if err != nil {
			return fmt.Errorf("failed to flush privileges: %w", err)
		}

		fmt.Printf("Privileges granted successfully to '%s'@'%s'\n", grantUser, host)
		return nil
	},
}

func init() {
	GrantCmd.Flags().StringVarP(&grantUser, "user", "u", "", "Username (required)")
	GrantCmd.Flags().StringVar(&grantHost, "host", "%", "User host (default: %)")
	GrantCmd.Flags().StringVarP(&grantPriv, "privileges", "p", "", "Privileges: SELECT, INSERT, UPDATE, DELETE, ALL, etc. (required)")
	GrantCmd.Flags().StringVarP(&grantDatabase, "database", "d", "*", "Database name")
	GrantCmd.Flags().StringVarP(&grantTable, "table", "t", "", "Table name (optional, if not specified applies to entire database)")
}
