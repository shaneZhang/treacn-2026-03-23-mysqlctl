package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var backupDatabase string
var backupOutput string

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup a database",
	Long:  `Backup a database to a SQL file using mysqldump. Requires mysqldump to be installed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		cfg := db.GetConfig()
		dbName := backupDatabase
		if dbName == "" {
			dbName = cfg.Database
		}
		if dbName == "" {
			return fmt.Errorf("database name is required (use --database)")
		}

		outputFile := backupOutput
		if outputFile == "" {
			timestamp := time.Now().Format("20060102_150405")
			outputFile = fmt.Sprintf("%s_%s.sql", dbName, timestamp)
		}

		mysqldumpPath, err := exec.LookPath("mysqldump")
		if err != nil {
			return fmt.Errorf("mysqldump not found. Please install MySQL client tools")
		}

		argsSlice := []string{
			"-h", cfg.Host,
			"-P", fmt.Sprintf("%d", cfg.Port),
			"-u", cfg.User,
		}
		if cfg.Password != "" {
			argsSlice = append(argsSlice, fmt.Sprintf("-p%s", cfg.Password))
		}
		argsSlice = append(argsSlice, dbName)
		argsSlice = append(argsSlice, []string{"-r", outputFile}...)

		mysqldumpCmd := exec.Command(mysqldumpPath, argsSlice...)
		mysqldumpCmd.Stdout = os.Stdout
		mysqldumpCmd.Stderr = os.Stderr

		err = mysqldumpCmd.Run()
		if err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}

		absPath, _ := filepath.Abs(outputFile)
		fmt.Printf("Database '%s' backed up to: %s\n", dbName, absPath)
		return nil
	},
}

func init() {
	BackupCmd.Flags().StringVarP(&backupDatabase, "database", "d", "", "Database name to backup")
	BackupCmd.Flags().StringVarP(&backupOutput, "output", "o", "", "Output file path")
}
