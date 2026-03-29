package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var (
	createIndexTable  string
	createIndexName   string
	createIndexCols   string
	createIndexUnique bool
)

var CreateIndexCmd = &cobra.Command{
	Use:   "create-index",
	Short: "Create an index on a table",
	Long:  `Create a new index on a table.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		if createIndexTable == "" {
			return fmt.Errorf("table name is required (use --table)")
		}
		if createIndexCols == "" {
			return fmt.Errorf("column(s) are required (use --columns)")
		}

		indexName := createIndexName
		if indexName == "" {
			indexName = fmt.Sprintf("idx_%s", createIndexCols)
		}

		var sqlQuery string
		if createIndexUnique {
			sqlQuery = fmt.Sprintf("CREATE UNIQUE INDEX `%s` ON `%s` (%s)",
				indexName, createIndexTable, createIndexCols)
		} else {
			sqlQuery = fmt.Sprintf("CREATE INDEX `%s` ON `%s` (%s)",
				indexName, createIndexTable, createIndexCols)
		}

		_, err := db.GetDB().Exec(sqlQuery)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}

		fmt.Printf("Index '%s' created successfully on table '%s'\n", indexName, createIndexTable)
		return nil
	},
}

func init() {
	CreateIndexCmd.Flags().StringVarP(&createIndexTable, "table", "t", "", "Table name (required)")
	CreateIndexCmd.Flags().StringVarP(&createIndexName, "name", "n", "", "Index name (optional, auto-generated if not provided)")
	CreateIndexCmd.Flags().StringVarP(&createIndexCols, "columns", "c", "", "Column(s): col1, col2, ... (required)")
	CreateIndexCmd.Flags().BoolVar(&createIndexUnique, "unique", false, "Create a unique index")
}
