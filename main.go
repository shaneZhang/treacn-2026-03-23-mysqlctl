package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mysqlctl/cmd"
)

var rootCmd = &cobra.Command{
	Use:   "mysqlctl",
	Short: "A MySQL command-line client tool",
	Long:  `A powerful MySQL command-line client built with Go and Cobra for database management and operations.`,
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
