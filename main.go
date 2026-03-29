package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mysqlctl/cmd"
	"mysqlctl/internal/db"
)

var (
	globalHost     string
	globalPort     int
	globalUser     string
	globalPassword string
	globalDatabase string
)

var rootCmd = &cobra.Command{
	Use:   "mysqlctl",
	Short: "A MySQL command-line client tool",
	Long:  `A powerful MySQL command-line client built with Go and Cobra for database management and operations.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "connect" || cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}
		if globalHost != "" && globalUser != "" {
			if !db.IsConnected() {
				err := db.Connect(globalHost, globalPort, globalUser, globalPassword, globalDatabase)
				if err != nil {
					return fmt.Errorf("auto-connect failed: %w", err)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&globalHost, "host", "H", "", "MySQL server host")
	rootCmd.PersistentFlags().IntVarP(&globalPort, "port", "P", 3306, "MySQL server port")
	rootCmd.PersistentFlags().StringVarP(&globalUser, "user", "u", "", "MySQL username")
	rootCmd.PersistentFlags().StringVarP(&globalPassword, "password", "p", "", "MySQL password")
	rootCmd.PersistentFlags().StringVarP(&globalDatabase, "database", "d", "", "Database name")
	rootCmd.PersistentFlags().Bool("help", false, "help for mysqlctl")
}

func main() {
	rootCmd.AddCommand(cmd.ConnectCmd)
	rootCmd.AddCommand(cmd.DisconnectCmd)
	rootCmd.AddCommand(cmd.StatusCmd)
	rootCmd.AddCommand(cmd.ShowDatabasesCmd)
	rootCmd.AddCommand(cmd.UseDatabaseCmd)
	rootCmd.AddCommand(cmd.CreateDatabaseCmd)
	rootCmd.AddCommand(cmd.DropDatabaseCmd)
	rootCmd.AddCommand(cmd.ShowTablesCmd)
	rootCmd.AddCommand(cmd.DescribeTableCmd)
	rootCmd.AddCommand(cmd.ShowCreateCmd)
	rootCmd.AddCommand(cmd.QueryCmd)
	rootCmd.AddCommand(cmd.SelectCmd)
	rootCmd.AddCommand(cmd.InsertCmd)
	rootCmd.AddCommand(cmd.UpdateCmd)
	rootCmd.AddCommand(cmd.DeleteCmd)
	rootCmd.AddCommand(cmd.ShowIndexCmd)
	rootCmd.AddCommand(cmd.CreateIndexCmd)
	rootCmd.AddCommand(cmd.ShowUsersCmd)
	rootCmd.AddCommand(cmd.CreateUserCmd)
	rootCmd.AddCommand(cmd.GrantCmd)
	rootCmd.AddCommand(cmd.ShowProcesslistCmd)
	rootCmd.AddCommand(cmd.ShowStatusCmd)
	rootCmd.AddCommand(cmd.BackupCmd)
	rootCmd.AddCommand(cmd.SourceCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
