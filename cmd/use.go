package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var UseDatabaseCmd = &cobra.Command{
	Use:   "use [database]",
	Short: "Switch to a database",
	Long:  `Select a database to use.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		dbName := args[0]
		_, err := db.GetDB().Exec(fmt.Sprintf("USE `%s`", dbName))
		if err != nil {
			return fmt.Errorf("failed to use database: %w", err)
		}

		// 更新配置文件中的数据库
		cfg, _ := db.LoadConfigFromFile()
		if cfg != nil {
			cfg.Database = dbName
			db.SaveConfigToFile(cfg)
		}

		fmt.Printf("Switched to database: %s\n", dbName)
		return nil
	},
}
