package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var (
	selectTable   string
	selectColumns string
	selectWhere   string
	selectOrderBy string
	selectLimit   int
	selectOffset  int
)

var SelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Query data with options",
	Long:  `Select data from a table with optional filters, ordering, and pagination.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		if selectTable == "" {
			return fmt.Errorf("table name is required (use --table)")
		}

		columns := "*"
		if selectColumns != "" {
			columns = selectColumns
		}

		sqlQuery := fmt.Sprintf("SELECT %s FROM %s", columns, escapeIdentifier(selectTable))

		if selectWhere != "" {
			sqlQuery += fmt.Sprintf(" WHERE %s", selectWhere)
		}

		if selectOrderBy != "" {
			sqlQuery += fmt.Sprintf(" ORDER BY %s", selectOrderBy)
		}

		if selectLimit > 0 {
			sqlQuery += fmt.Sprintf(" LIMIT %d", selectLimit)
		}

		if selectOffset > 0 {
			sqlQuery += fmt.Sprintf(" OFFSET %d", selectOffset)
		}

		rows, err := db.GetDB().Query(sqlQuery)
		if err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()

		printResult(rows)
		return nil
	},
}

func init() {
	SelectCmd.Flags().StringVarP(&selectTable, "table", "t", "", "Table name (required)")
	SelectCmd.Flags().StringVarP(&selectColumns, "columns", "c", "*", "Columns to select")
	SelectCmd.Flags().StringVarP(&selectWhere, "where", "w", "", "WHERE condition")
	SelectCmd.Flags().StringVarP(&selectOrderBy, "order-by", "o", "", "ORDER BY clause")
	SelectCmd.Flags().IntVarP(&selectLimit, "limit", "l", 0, "Limit number of rows")
	SelectCmd.Flags().IntVarP(&selectOffset, "offset", "s", 0, "Offset for pagination")
}
