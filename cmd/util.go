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

	// 使用 sql.RawBytes 来接收数据
	values := make([]sql.RawBytes, len(columns))
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
				fmt.Print(string(v))
			}
		}
		fmt.Println()
	}
}

func checkConnection() bool {
	if !db.IsConnected() {
		// 尝试从配置文件自动连接
		if err := db.AutoConnect(); err != nil {
			fmt.Println("Not connected to MySQL. Use 'mysqlctl connect' first.")
			return false
		}
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
