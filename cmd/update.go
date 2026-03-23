package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var (
	updateTable string
	updateSet   string
	updateWhere string
)

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update data in a table",
	Long:  `Update existing rows in a table.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		if updateTable == "" {
			return fmt.Errorf("table name is required (use --table)")
		}

		if updateSet == "" {
			return fmt.Errorf("set clause is required (use --set)")
		}

		sqlQuery := fmt.Sprintf("UPDATE %s SET %s", escapeIdentifier(updateTable), updateSet)

		if updateWhere != "" {
			sqlQuery += fmt.Sprintf(" WHERE %s", updateWhere)
		}

		result, err := db.GetDB().Exec(sqlQuery)
		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		printAffectedRows(result)
		return nil
	},
}

func init() {
	UpdateCmd.Flags().StringVarP(&updateTable, "table", "t", "", "Table name (required)")
	UpdateCmd.Flags().StringVarP(&updateSet, "set", "s", "", "Column=value pairs: col1='val1', col2='val2' (required)")
	UpdateCmd.Flags().StringVarP(&updateWhere, "where", "w", "", "WHERE condition")
}
