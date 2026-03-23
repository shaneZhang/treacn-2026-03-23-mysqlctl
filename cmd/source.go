package cmd

import (
	"fmt"
	"os"
	"strings"
	"unicode"

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

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		statements := splitSQLStatements(string(content))

		fmt.Printf("Executing %d statements...\n", len(statements))

		for i, stmt := range statements {
			_, err := db.GetDB().Exec(stmt)
			if err != nil {
				fmt.Printf("Error executing statement %d: %v\n", i+1, err)
				fmt.Printf("Statement: %.100s\n", stmt)
				return fmt.Errorf("failed to execute statement: %w", err)
			}
		}

		fmt.Println("All statements executed successfully")
		return nil
	},
}

func splitSQLStatements(sql string) []string {
	var statements []string
	var currentStmt strings.Builder
	var inString bool
	var inComment bool
	var stringChar rune
	var delimiter string = ";"

	lines := strings.Split(sql, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "--") || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		if strings.HasPrefix(strings.ToUpper(trimmedLine), "DELIMITER") {
			parts := strings.Fields(trimmedLine)
			if len(parts) >= 2 {
				delimiter = parts[1]
			}
			continue
		}

		var inMultiLineComment bool
		i := 0
		runes := []rune(line)
		lineLen := len(runes)

		for i < lineLen {
			r := runes[i]

			if !inString && !inComment && !inMultiLineComment && i+1 < lineLen && runes[i] == '/' && runes[i+1] == '*' {
				inMultiLineComment = true
				i += 2
				continue
			}

			if inMultiLineComment && i+1 < lineLen && runes[i] == '*' && runes[i+1] == '/' {
				inMultiLineComment = false
				i += 2
				continue
			}

			if inMultiLineComment {
				i++
				continue
			}

			if !inString && i+1 < lineLen && runes[i] == '/' && runes[i+1] == '/' {
				break
			}

			if (r == '\'' || r == '"') && !inComment {
				if !inString {
					inString = true
					stringChar = r
				} else if r == stringChar {
					if i+1 >= lineLen || runes[i+1] != stringChar {
						inString = false
					} else {
						i++
					}
				}
			}

			if !inString && !inComment {
				delimiterFound := false
				if i+len(delimiter) <= lineLen {
					if string(runes[i:i+len(delimiter)]) == delimiter {
						stmt := strings.TrimSpace(currentStmt.String())
						if stmt != "" {
							statements = append(statements, stmt)
						}
						currentStmt.Reset()
						i += len(delimiter)
						delimiterFound = true
					}
				}
				if delimiterFound {
					continue
				}
			}

			if !inComment || !unicode.IsSpace(r) {
				currentStmt.WriteRune(r)
			}
			i++
		}

		if currentStmt.Len() > 0 {
			currentStmt.WriteRune('\n')
		}
	}

	stmt := strings.TrimSpace(currentStmt.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

func init() {
	SourceCmd.Flags().StringVarP(&sourceFile, "file", "f", "", "SQL file to execute")
}
