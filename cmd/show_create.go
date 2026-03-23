package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var ShowCreateCmd = &cobra.Command{
	Use:   "show-create [table]",
	Short: "Show CREATE TABLE statement",
	Long:  `Display the CREATE TABLE statement used to create a table.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		tableName := args[0]
		rows, err := db.GetDB().Query(fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName))
		if err != nil {
			return fmt.Errorf("failed to show create table: %w", err)
		}
		defer rows.Close()

		printResult(rows)
		return nil
	},
}
