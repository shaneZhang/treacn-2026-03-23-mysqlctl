package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"mysqlctl/internal/db"
)

func escapeIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func escapeValue(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
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
				switch val := v.(type) {
				case []byte:
					fmt.Print(string(val))
				case string:
					fmt.Print(val)
				case int, int8, int16, int32, int64:
					fmt.Print(val)
				case uint, uint8, uint16, uint32, uint64:
					fmt.Print(val)
				case float32, float64:
					fmt.Print(val)
				default:
					fmt.Print(val)
				}
			}
		}
		fmt.Println()
	}
}

func checkConnection() bool {
	if db.IsConnected() {
		return true
	}

	cfg, err := db.LoadConfig()
	if err != nil {
		host := os.Getenv("MYSQL_HOST")
		if host != "" {
			port := 3306
			if p := os.Getenv("MYSQL_PORT"); p != "" {
				fmt.Sscanf(p, "%d", &port)
			}
			user := os.Getenv("MYSQL_USER")
			if user == "" {
				user = "root"
			}
			password := os.Getenv("MYSQL_PASSWORD")
			database := os.Getenv("MYSQL_DATABASE")

			err = db.Connect(host, port, user, password, database)
			if err != nil && database != "" {
				err = db.Connect(host, port, user, password, "")
				if err != nil {
					fmt.Printf("Failed to connect: %v\n", err)
					return false
				}
				return true
			} else if err != nil {
				fmt.Printf("Failed to connect: %v\n", err)
				return false
			}
			return true
		}

		fmt.Println("Not connected to MySQL. Use 'mysqlctl connect' first.")
		return false
	}

	err = db.Connect(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)
	if err != nil && cfg.Database != "" {
		err = db.Connect(cfg.Host, cfg.Port, cfg.User, cfg.Password, "")
		if err != nil {
			fmt.Printf("Failed to reconnect: %v\n", err)
			fmt.Println("Use 'mysqlctl connect' to establish a new connection.")
			return false
		}
		return true
	} else if err != nil {
		fmt.Printf("Failed to reconnect: %v\n", err)
		fmt.Println("Use 'mysqlctl connect' to establish a new connection.")
		return false
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
