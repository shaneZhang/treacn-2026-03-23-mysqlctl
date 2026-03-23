package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var (
	deleteTable string
	deleteWhere string
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete data from a table",
	Long:  `Delete rows from a table.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		if deleteTable == "" {
			return fmt.Errorf("table name is required (use --table)")
		}

		if deleteWhere == "" {
			return fmt.Errorf("where clause is required (use --where)")
		}

		// Validate table name to prevent SQL injection
		if !isValidIdentifier(deleteTable) {
			return fmt.Errorf("invalid table name: contains potentially dangerous characters")
		}

		sqlQuery := fmt.Sprintf("DELETE FROM `%s` WHERE %s", deleteTable, deleteWhere)

		result, err := db.GetDB().Exec(sqlQuery)
		if err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}

		printAffectedRows(result)
		return nil
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&deleteTable, "table", "t", "", "Table name (required)")
	DeleteCmd.Flags().StringVarP(&deleteWhere, "where", "w", "", "WHERE condition (required)")
}
