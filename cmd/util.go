package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"mysqlctl/internal/db"
)

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
		fmt.Println("Not connected to MySQL. Use 'mysqlctl connect' first.")
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
