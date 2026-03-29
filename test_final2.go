package main

import (
	"fmt"
	"strings"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type TestResult struct {
	Category string
	Name     string
	Passed   bool
	Message  string
}

var results []TestResult
var testDB = "mysqlctl_final2_test"

func addResult(category, name string, passed bool, message string) {
	results = append(results, TestResult{category, name, passed, message})
}

func getDB(database string) (*sql.DB, error) {
	dsn := fmt.Sprintf("zhangyuqing:zhangyuqing@tcp(192.168.31.210:3306)/%s?parseTime=true", database)
	return sql.Open("mysql", dsn)
}

func main() {
	fmt.Println("=== MySQLCTL 最终测试报告 ===")
	
	// 先创建测试数据库
	db, err := getDB("")
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE `%s`", testDB))
	if err != nil {
		fmt.Printf("创建测试库失败: %v\n", err)
		return
	}
	db.Close()
	fmt.Println("✅ 准备测试环境完成")
	
	// 使用指定数据库重新连接
	db, err = getDB(testDB)
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer db.Close()
	defer func() {
		db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))
	}()
	
	// ==================== 1. 连接测试 ====================
	fmt.Println("\n=== 1. 基础连接测试 ===")
	if err := db.Ping(); err != nil {
		addResult("连接", "基本连接测试", false, err.Error())
		fmt.Printf("  ❌ 连接失败: %v\n", err)
	} else {
		addResult("连接", "基本连接测试", true, "")
		fmt.Println("  ✅ 数据库连接正常")
	}
	
	// ==================== 2. 表操作 ====================
	fmt.Println("\n=== 2. 表操作测试 ===")
	
	_, err = db.Exec(`
		CREATE TABLE test_users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(50) NOT NULL,
			email VARCHAR(100) UNIQUE,
			age INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		addResult("表操作", "CREATE TABLE", false, err.Error())
		fmt.Printf("  ❌ 创建表失败: %v\n", err)
	} else {
		addResult("表操作", "CREATE TABLE", true, "")
		fmt.Println("  ✅ CREATE TABLE 成功")
	}
	
	rows, _ := db.Query("SHOW TABLES")
	defer rows.Close()
	var tableName string
	if rows.Next() && rows.Scan(&tableName) == nil && tableName == "test_users" {
		addResult("表操作", "SHOW TABLES", true, "")
		fmt.Println("  ✅ SHOW TABLES 成功")
	} else {
		addResult("表操作", "SHOW TABLES", false, "表不存在")
		fmt.Println("  ❌ SHOW TABLES 失败")
	}
	
	// ==================== 3. 数据操作 ====================
	fmt.Println("\n=== 3. 数据操作测试 ===")
	
	// INSERT
	result, err := db.Exec(`
		INSERT INTO test_users (name, email, age) VALUES (?, ?, ?)
	`, "张三", "zhangsan@test.com", 25)
	if err != nil {
		addResult("数据操作", "INSERT", false, err.Error())
		fmt.Printf("  ❌ INSERT失败: %v\n", err)
	} else {
		id, _ := result.LastInsertId()
		addResult("数据操作", "INSERT", true, fmt.Sprintf("ID:%d", id))
		fmt.Printf("  ✅ INSERT成功 (新记录ID:%d)\n", id)
	}
	
	// SELECT
	var name string
	var age int
	err = db.QueryRow("SELECT name, age FROM test_users WHERE id = 1").Scan(&name, &age)
	if err != nil {
		addResult("数据操作", "SELECT", false, err.Error())
		fmt.Printf("  ❌ SELECT失败: %v\n", err)
	} else {
		addResult("数据操作", "SELECT", true, fmt.Sprintf("name:%s,age:%d", name, age))
		fmt.Printf("  ✅ SELECT成功 (name:%s, age:%d)\n", name, age)
	}
	
	// UPDATE
	result, err = db.Exec("UPDATE test_users SET age = ? WHERE name = ?", 26, "张三")
	if err != nil {
		addResult("数据操作", "UPDATE", false, err.Error())
		fmt.Printf("  ❌ UPDATE失败: %v\n", err)
	} else {
		affected, _ := result.RowsAffected()
		addResult("数据操作", "UPDATE", true, fmt.Sprintf("影响%d行", affected))
		fmt.Printf("  ✅ UPDATE成功 (影响%d行)\n", affected)
	}
	
	// DELETE
	result, err = db.Exec("DELETE FROM test_users WHERE name = ?", "张三")
	if err != nil {
		addResult("数据操作", "DELETE", false, err.Error())
		fmt.Printf("  ❌ DELETE失败: %v\n", err)
	} else {
		affected, _ := result.RowsAffected()
		addResult("数据操作", "DELETE", true, fmt.Sprintf("删除%d行", affected))
		fmt.Printf("  ✅ DELETE成功 (删除%d行)\n", affected)
	}
	
	// 批量插入
	for i := 0; i < 5; i++ {
		db.Exec("INSERT INTO test_users (name, email, age) VALUES (?, ?, ?)", 
			fmt.Sprintf("用户%d", i), fmt.Sprintf("user%d@test.com", i), 20+i)
	}
	
	// WHERE条件
	rows, err = db.Query("SELECT name FROM test_users WHERE age > ?", 22)
	if err != nil {
		addResult("数据操作", "WHERE条件", false, err.Error())
	} else {
		defer rows.Close()
		names := []string{}
		for rows.Next() {
			var n string
			rows.Scan(&n)
			names = append(names, n)
		}
		addResult("数据操作", "WHERE条件", true, fmt.Sprintf("匹配:%d个", len(names)))
		fmt.Printf("  ✅ WHERE条件查询 (age>22, 匹配:%d个)\n", len(names))
	}
	
	// ORDER BY + LIMIT
	rows, err = db.Query("SELECT name, age FROM test_users ORDER BY age DESC LIMIT 3")
	if err != nil {
		addResult("数据操作", "ORDER BY + LIMIT", false, err.Error())
	} else {
		defer rows.Close()
		if rows.Next() {
			var n string
			var a int
			rows.Scan(&n, &a)
			addResult("数据操作", "ORDER BY + LIMIT", true, fmt.Sprintf("首行:%s,%d", n, a))
			fmt.Printf("  ✅ ORDER BY + LIMIT (首行:%s-%d)\n", n, a)
		}
	}
	
	// LIKE
	rows, err = db.Query("SELECT name FROM test_users WHERE name LIKE ?", "%3%")
	if err != nil {
		addResult("数据操作", "LIKE模糊查询", false, err.Error())
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("数据操作", "LIKE模糊查询", true, fmt.Sprintf("匹配:%d个", count))
		fmt.Printf("  ✅ LIKE模糊查询 (含'3':%d个)\n", count)
	}
	
	// ==================== 4. 索引操作 ====================
	fmt.Println("\n=== 4. 索引操作测试 ===")
	
	// 创建普通索引
	_, err = db.Exec("CREATE INDEX idx_name ON test_users(name)")
	if err != nil {
		addResult("索引", "CREATE INDEX", false, err.Error())
		fmt.Printf("  ❌ 创建普通索引失败: %v\n", err)
	} else {
		addResult("索引", "CREATE INDEX", true, "")
		fmt.Println("  ✅ CREATE INDEX (普通索引) 成功")
	}
	
	// 创建唯一索引
	_, err = db.Exec("CREATE UNIQUE INDEX idx_email_unique ON test_users(email)")
	if err != nil {
		addResult("索引", "CREATE UNIQUE INDEX", false, err.Error())
		fmt.Printf("  ❌ 创建唯一索引失败: %v\n", err)
	} else {
		addResult("索引", "CREATE UNIQUE INDEX", true, "")
		fmt.Println("  ✅ CREATE UNIQUE INDEX (唯一索引) 成功")
	}
	
	// 查看索引
	rows, err = db.Query("SHOW INDEX FROM test_users")
	if err != nil {
		addResult("索引", "SHOW INDEX", false, err.Error())
		fmt.Printf("  ❌ 查看索引失败: %v\n", err)
	} else {
		defer rows.Close()
		indexes := make(map[string]bool)
		for rows.Next() {
			var dummy interface{}
			var table, nonUnique, keyName string
			rows.Scan(&table, &nonUnique, &keyName, &dummy, &dummy, &dummy, &dummy, &dummy, &dummy, &dummy, &dummy, &dummy, &dummy, &dummy, &dummy, &dummy)
			indexes[keyName] = true
		}
		idxList := []string{}
		for k := range indexes {
			idxList = append(idxList, k)
		}
		addResult("索引", "SHOW INDEX", true, fmt.Sprintf("索引:%v", idxList))
		fmt.Printf("  ✅ SHOW INDEX (索引列表:%v)\n", idxList)
	}
	
	// ==================== 5. 系统信息查询 ====================
	fmt.Println("\n=== 5. 系统信息查询 ===")
	
	// 创建另一个连接用于系统查询
	dbGlobal, _ := getDB("mysql")
	defer dbGlobal.Close()
	
	rows, _ = dbGlobal.Query("SELECT user, host FROM mysql.user LIMIT 3")
	if rows != nil {
		defer rows.Close()
		users := []string{}
		for rows.Next() {
			var u, h string
			rows.Scan(&u, &h)
			users = append(users, fmt.Sprintf("%s@%s", u, h))
		}
		addResult("系统查询", "列出用户", true, fmt.Sprintf("%d个用户", len(users)))
		fmt.Printf("  ✅ 列出MySQL用户 (前3个:%v)\n", users)
	}
	
	// SHOW STATUS
	var statusName, statusValue string
	err = dbGlobal.QueryRow("SHOW GLOBAL STATUS LIKE 'Uptime'").Scan(&statusName, &statusValue)
	if err != nil {
		addResult("系统查询", "服务器状态", false, err.Error())
	} else {
		addResult("系统查询", "服务器状态", true, fmt.Sprintf("运行:%s秒", statusValue))
		fmt.Printf("  ✅ 服务器状态 (运行时间:%s秒)\n", statusValue)
	}
	
	// ==================== 6. 错误处理 ====================
	fmt.Println("\n=== 6. 错误处理测试 ===")
	
	// 不存在的表
	_, err = db.Query("SELECT * FROM table_that_not_exist_123")
	if err != nil {
		addResult("错误处理", "查询不存在的表", true, "正确报错")
		fmt.Println("  ✅ 查询不存在的表: 正确报错")
	} else {
		addResult("错误处理", "查询不存在的表", false, "未报错")
		fmt.Println("  ❌ 查询不存在的表: 应该报错但未报错")
	}
	
	// SQL语法错误
	_, err = db.Query("SELECT * FROM test_users WHERE COLUMN_NOT_EXIST = 1")
	if err != nil {
		addResult("错误处理", "字段不存在", true, "正确报错")
		fmt.Println("  ✅ 查询不存在的字段: 正确报错")
	} else {
		addResult("错误处理", "字段不存在", false, "未报错")
		fmt.Println("  ❌ 查询不存在的字段: 应该报错但未报错")
	}
	
	// 唯一约束重复
	_, err = db.Exec("INSERT INTO test_users (name, email) VALUES ('test', 'user0@test.com')")
	if err != nil {
		addResult("错误处理", "唯一约束冲突", true, "正确报错")
		fmt.Println("  ✅ 唯一约束冲突: 正确报错")
	} else {
		addResult("错误处理", "唯一约束冲突", false, "未报错")
		fmt.Println("  ❌ 唯一约束冲突: 应该报错但未报错")
	}
	
	// 特殊字符
	_, err = db.Exec(
		"INSERT INTO test_users (name, email) VALUES (?, ?)",
		`Name 'with' "quotes" \and\slashes`, "special@test.com",
	)
	if err != nil {
		addResult("错误处理", "特殊字符转义", false, err.Error())
		fmt.Printf("  ❌ 特殊字符处理失败: %v\n", err)
	} else {
		addResult("错误处理", "特殊字符转义", true, "")
		fmt.Println("  ✅ 特殊字符转义处理成功")
	}
	
	// ==================== 7. 生成报告 ====================
	generateReport()
}

func generateReport() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("                        测 试 报 告")
	fmt.Println(strings.Repeat("=", 70))
	
	total := len(results)
	passed := 0
	categoryStats := make(map[string]struct{ total, passed int })
	
	for _, r := range results {
		stats := categoryStats[r.Category]
		stats.total++
		if r.Passed {
			passed++
			stats.passed++
		}
		categoryStats[r.Category] = stats
	}
	
	fmt.Printf("\n📊 测试总览\n")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("  总测试用例: %d\n", total)
	fmt.Printf("  通    过: %d  (%.2f%%)\n", passed, float64(passed)*100/float64(total))
	fmt.Printf("  失    败: %d\n", total - passed)
	
	fmt.Printf("\n📈 按类别统计\n")
	fmt.Println(strings.Repeat("-", 70))
	for cat, stats := range categoryStats {
		rate := float64(stats.passed) * 100 / float64(stats.total)
		fmt.Printf("  %-12s %2d/%2d  (%.1f%%)\n", cat, stats.passed, stats.total, rate)
	}
	
	failed := []TestResult{}
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}
	
	if len(failed) > 0 {
		fmt.Printf("\n❌ 失败详情\n")
		fmt.Println(strings.Repeat("-", 70))
		for _, r := range failed {
			fmt.Printf("  [%s] %-25s → %s\n", r.Category, r.Name, r.Message)
		}
	}
	
	fmt.Printf("\n✅ 通过的测试用例\n")
	fmt.Println(strings.Repeat("-", 70))
	for _, r := range results {
		if r.Passed {
			fmt.Printf("  ✓ [%s] %s", r.Category, r.Name)
			if r.Message != "" {
				fmt.Printf(" (%s)", r.Message)
			}
			fmt.Println()
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 70))
}
