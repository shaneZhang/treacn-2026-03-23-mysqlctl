package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var showStatusExtended bool

var ShowStatusCmd = &cobra.Command{
	Use:   "show-status",
	Short: "Show MySQL server status",
	Long:  `Display MySQL server status variables. Use --extended to show all variables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		var sqlQuery string
		if showStatusExtended {
			sqlQuery = "SHOW GLOBAL STATUS"
		} else {
			sqlQuery = "SHOW GLOBAL STATUS LIKE 'Threads_connected%'"
		}

		rows, err := db.GetDB().Query(sqlQuery)
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}
		defer rows.Close()

		printResult(rows)
		return nil
	},
}

func init() {
	ShowStatusCmd.Flags().BoolVarP(&showStatusExtended, "extended", "e", false, "Show all status variables")
}
