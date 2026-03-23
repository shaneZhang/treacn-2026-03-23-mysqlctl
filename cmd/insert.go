package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var (
	insertTable  string
	insertValues string
	insertSet    string
)

var InsertCmd = &cobra.Command{
	Use:   "insert",
	Short: "Insert data into a table",
	Long:  `Insert a new row into a table. Use --values for simple value insertion or --set for column=value syntax.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		if insertTable == "" {
			return fmt.Errorf("table name is required (use --table)")
		}

		var sqlQuery string
		if insertSet != "" {
			sqlQuery = fmt.Sprintf("INSERT INTO %s SET %s", escapeIdentifier(insertTable), insertSet)
		} else if insertValues != "" {
			sqlQuery = fmt.Sprintf("INSERT INTO %s VALUES (%s)", escapeIdentifier(insertTable), insertValues)
		} else {
			return fmt.Errorf("either --values or --set must be specified")
		}

		result, err := db.GetDB().Exec(sqlQuery)
		if err != nil {
			return fmt.Errorf("insert failed: %w", err)
		}

		printAffectedRows(result)
		return nil
	},
}

func init() {
	InsertCmd.Flags().StringVarP(&insertTable, "table", "t", "", "Table name (required)")
	InsertCmd.Flags().StringVarP(&insertValues, "values", "v", "", "Values in format: 'value1', 'value2', ...")
	InsertCmd.Flags().StringVarP(&insertSet, "set", "s", "", "Column=value pairs: col1='val1', col2='val2'")
}
