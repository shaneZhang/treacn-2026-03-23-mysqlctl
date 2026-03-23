package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"mysqlctl/internal/db"
)

// isValidIdentifier checks if a string is a valid MySQL identifier
// Only allows alphanumeric characters, underscores, and dollar signs
// Also allows '%' and '*' for wildcard patterns in user hosts and grants
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	// Allow '%' as wildcard for hosts and grants
	if s == "%" || s == "*" {
		return true
	}
	// MySQL identifiers can contain alphanumeric, underscore, and dollar sign
	// They cannot start with a digit (unless quoted, but we use backticks)
	matched, _ := regexp.MatchString("^[a-zA-Z_][a-zA-Z0-9_$]*$", s)
	return matched
}

// isValidHost checks if a string is a valid MySQL user host
// Allows IP addresses, hostnames, and wildcards like %
func isValidHost(s string) bool {
	if s == "" {
		return false
	}
	// Allow '%' as wildcard for hosts
	if s == "%" {
		return true
	}
	// Allow IP addresses (IPv4 and IPv6)
	if matched, _ := regexp.MatchString(`^[0-9a-fA-F.:]+$`, s); matched {
		return true
	}
	// Allow hostnames with wildcards
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_%\-\.]+$`, s)
	return matched
}

// isValidColumnList checks if a string is a valid comma-separated column list
func isValidColumnList(s string) bool {
	if s == "" {
		return false
	}
	// Allow comma-separated identifiers, possibly with spaces
	// Also allow * for all columns
	if s == "*" {
		return true
	}
	// Split by comma and validate each part
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Allow simple identifiers or table.column format
		if !isValidIdentifier(part) {
			// Check for table.column format
			if strings.Contains(part, ".") {
				subParts := strings.Split(part, ".")
				if len(subParts) == 2 {
					if isValidIdentifier(subParts[0]) && isValidIdentifier(subParts[1]) {
						continue
					}
				}
			}
			return false
		}
	}
	return true
}

// isValidPrivilege checks if a privilege string is valid
func isValidPrivilege(s string) bool {
	if s == "" {
		return false
	}
	// Allow common privilege names and comma-separated list
	// Also allow ALL PRIVILEGES
	allowed := map[string]bool{
		"ALL": true, "ALL PRIVILEGES": true, "SELECT": true, "INSERT": true,
		"UPDATE": true, "DELETE": true, "CREATE": true, "DROP": true,
		"RELOAD": true, "SHUTDOWN": true, "PROCESS": true, "FILE": true,
		"REFERENCES": true, "INDEX": true, "ALTER": true, "SHOW DATABASES": true,
		"SUPER": true, "CREATE TEMPORARY TABLES": true, "LOCK TABLES": true,
		"EXECUTE": true, "REPLICATION SLAVE": true, "REPLICATION CLIENT": true,
		"CREATE VIEW": true, "SHOW VIEW": true, "CREATE ROUTINE": true,
		"ALTER ROUTINE": true, "CREATE USER": true, "EVENT": true, "TRIGGER": true,
		"CREATE TABLESPACE": true,
	}

	// Check if it's a single known privilege
	if allowed[strings.ToUpper(s)] {
		return true
	}

	// Check comma-separated privileges
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if !allowed[strings.ToUpper(part)] {
			// Check for column-level privileges like SELECT(col1, col2)
			if idx := strings.Index(part, "("); idx > 0 {
				privName := strings.TrimSpace(part[:idx])
				if allowed[strings.ToUpper(privName)] {
					// Validate column list inside parentheses
					colsPart := part[idx+1:]
					if idx2 := strings.LastIndex(colsPart, ")"); idx2 > 0 {
						colsPart = colsPart[:idx2]
						if isValidColumnList(colsPart) {
							continue
						}
					}
				}
			}
			return false
		}
	}
	return true
}

func printResult(rows *sql.Rows) {
	if rows == nil {
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting columns: %v\n", err)
		return
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	fmt.Println(strings.Join(columns, "\t"))
	fmt.Println(strings.Repeat("-", 80))

	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
			return
		}
		for i, v := range values {
			if i > 0 {
				fmt.Print("\t")
			}
			if v == nil {
				fmt.Print("NULL")
			} else {
				fmt.Print(v)
			}
		}
		fmt.Println()
	}
}

func checkConnection() bool {
	if !db.IsConnected() {
		// Try to reconnect using saved config
		err := db.ConnectWithSavedConfig()
		if err != nil {
			fmt.Println("Not connected to MySQL. Use 'mysqlctl connect' first.")
			return false
		}
		// Reconnection successful
		return true
	}
	return true
}

func printAffectedRows(result sql.Result) {
	affected, err := result.RowsAffected()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting affected rows: %v\n", err)
		return
	}
	fmt.Printf("Affected rows: %d\n", affected)
}
