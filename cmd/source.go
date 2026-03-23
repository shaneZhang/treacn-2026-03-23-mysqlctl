package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"mysqlctl/internal/db"
)

var sourceFile string

var SourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Execute SQL from a file",
	Long:  `Execute SQL statements from a file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !checkConnection() {
			return nil
		}

		filePath := sourceFile
		if filePath == "" {
			filePath = args[0]
		}

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		var statements []string
		var currentStatement strings.Builder

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(strings.TrimSpace(line), "--") {
				continue
			}

			currentStatement.WriteString(line)
			currentStatement.WriteString("\n")

			if strings.Contains(line, ";") {
				stmt := strings.TrimSpace(currentStatement.String())
				if stmt != "" && stmt != ";" {
					statements = append(statements, stmt)
				}
				currentStatement.Reset()
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}

		if currentStatement.Len() > 0 {
			stmt := strings.TrimSpace(currentStatement.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
		}

		fmt.Printf("Executing %d statements...\n", len(statements))

		for i, stmt := range statements {
			_, err := db.GetDB().Exec(stmt)
			if err != nil {
				fmt.Printf("Error executing statement %d: %v\n", i+1, err)
				fmt.Printf("Statement: %s\n", stmt)
				return fmt.Errorf("failed to execute statement: %w", err)
			}
		}

		fmt.Println("All statements executed successfully")
		return nil
	},
}

func init() {
	SourceCmd.Flags().StringVarP(&sourceFile, "file", "f", "", "SQL file to execute")
}
