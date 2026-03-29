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
var testDB = "mysqlctl_single_test"

func addResult(category, name string, passed bool, message string) {
	results = append(results, TestResult{category, name, passed, message})
}

func main() {
	fmt.Println("=== MySQLCTL 核心功能测试 (直接使用 SQL驱动) ===")
	
	dsn := "zhangyuqing:zhangyuqing@tcp(192.168.31.210:3306)/?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer db.Close()
	
	if err := db.Ping(); err != nil {
		fmt.Printf("Ping失败: %v\n", err)
		return
	}
	
	fmt.Println("✅ 成功连接到 MySQL 服务器")
	
	// 清理
	db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))
	
	// ==================== 1. 数据库操作 ====================
	fmt.Println("\n=== 1. 数据库操作测试 ===")
	
	// 创建数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE `%s`", testDB))
	if err != nil {
		addResult("数据库", "创建数据库", false, err.Error())
		fmt.Printf("  ❌ 创建数据库失败: %v\n", err)
	} else {
		addResult("数据库", "创建数据库", true, "")
		fmt.Println("  ✅ 创建数据库成功")
	}
	
	// 列出数据库
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		addResult("数据库", "列出数据库", false, err.Error())
		fmt.Printf("  ❌ 列出数据库失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		found := false
		for rows.Next() {
			var name string
			rows.Scan(&name)
			count++
			if name == testDB {
				found = true
			}
		}
		addResult("数据库", "列出数据库", true, fmt.Sprintf("共%d个", count))
		fmt.Printf("  ✅ 列出数据库成功(共%d个, 测试库存在:%v)\n", count, found)
	}
	
	// 切换到测试数据库
	_, err = db.Exec(fmt.Sprintf("USE `%s`", testDB))
	if err != nil {
		fmt.Printf("切换数据库失败: %v\n", err)
		return
	}
	fmt.Println("  ✅ 切换到测试数据库")
	
	// ==================== 2. 表操作 ====================
	fmt.Println("\n=== 2. 表操作测试 ===")
	
	// 创建表
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
		addResult("表操作", "创建表", false, err.Error())
		fmt.Printf("  ❌ 创建表失败: %v\n", err)
	} else {
		addResult("表操作", "创建表", true, "")
		fmt.Println("  ✅ 创建表成功")
	}
	
	// 列出表
	rows, err = db.Query("SHOW TABLES")
	if err != nil {
		addResult("表操作", "列出表", false, err.Error())
		fmt.Printf("  ❌ 列出表失败: %v\n", err)
	} else {
		defer rows.Close()
		tables := []string{}
		for rows.Next() {
			var name string
			rows.Scan(&name)
			tables = append(tables, name)
		}
		addResult("表操作", "列出表", true, fmt.Sprintf("%v", tables))
		fmt.Printf("  ✅ 列出表成功: %v\n", tables)
	}
	
	// 查看表结构
	rows, err = db.Query("DESCRIBE test_users")
	if err != nil {
		addResult("表操作", "查看表结构", false, err.Error())
		fmt.Printf("  ❌ 查看表结构失败: %v\n", err)
	} else {
		defer rows.Close()
		cols := []string{}
		for rows.Next() {
			var field, typ, null, key, dflt, extra string
			rows.Scan(&field, &typ, &null, &key, &dflt, &extra)
			cols = append(cols, field)
		}
		addResult("表操作", "查看表结构", true, fmt.Sprintf("列:%v", cols))
		fmt.Printf("  ✅ 查看表结构成功: %v\n", cols)
	}
	
	// ==================== 3. 数据操作 ====================
	fmt.Println("\n=== 3. 数据操作测试 ===")
	
	// INSERT
	result, err := db.Exec(`
		INSERT INTO test_users (name, email, age) VALUES (?, ?, ?)
	`, "张三", "zhangsan@example.com", 25)
	if err != nil {
		addResult("数据操作", "INSERT", false, err.Error())
		fmt.Printf("  ❌ INSERT失败: %v\n", err)
	} else {
		id, _ := result.LastInsertId()
		addResult("数据操作", "INSERT", true, fmt.Sprintf("ID:%d", id))
		fmt.Printf("  ✅ INSERT成功 (ID:%d)\n", id)
	}
	
	// SELECT
	rows, err = db.Query("SELECT id, name, email FROM test_users WHERE name = ?", "张三")
	if err != nil {
		addResult("数据操作", "SELECT", false, err.Error())
		fmt.Printf("  ❌ SELECT失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			var id int
			var name, email string
			rows.Scan(&id, &name, &email)
			count++
		}
		addResult("数据操作", "SELECT", true, fmt.Sprintf("返回%d行", count))
		fmt.Printf("  ✅ SELECT成功 (返回%d行)\n", count)
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
	
	// 批量插入测试
	users := []struct{ name, email string; age int }{
		{"用户A", "a@test.com", 20},
		{"用户B", "b@test.com", 25},
		{"用户C", "c@test.com", 30},
	}
	for _, u := range users {
		db.Exec("INSERT INTO test_users (name, email, age) VALUES (?, ?, ?)", u.name, u.email, u.age)
	}
	
	// WHERE条件
	rows, err = db.Query("SELECT name FROM test_users WHERE age > ?", 22)
	if err != nil {
		addResult("数据操作", "WHERE条件", false, err.Error())
	} else {
		defer rows.Close()
		names := []string{}
		for rows.Next() {
			var name string
			rows.Scan(&name)
			names = append(names, name)
		}
		addResult("数据操作", "WHERE条件", true, fmt.Sprintf("匹配:%v", names))
		fmt.Printf("  ✅ WHERE条件查询 (age>22): %v\n", names)
	}
	
	// ORDER BY
	rows, err = db.Query("SELECT name, age FROM test_users ORDER BY age DESC LIMIT 2")
	if err != nil {
		addResult("数据操作", "ORDER BY", false, err.Error())
	} else {
		defer rows.Close()
		var name string
		var age int
		if rows.Next() {
			rows.Scan(&name, &age)
			addResult("数据操作", "ORDER BY", true, fmt.Sprintf("第一行:%s(%d)", name, age))
			fmt.Printf("  ✅ ORDER BY排序 (按age降序, 第一个:%s-%d)\n", name, age)
		}
	}
	
	// LIKE模糊查询
	rows, err = db.Query("SELECT name FROM test_users WHERE name LIKE ?", "%户%")
	if err != nil {
		addResult("数据操作", "LIKE模糊查询", false, err.Error())
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("数据操作", "LIKE模糊查询", true, fmt.Sprintf("匹配%d行", count))
		fmt.Printf("  ✅ LIKE模糊查询 (含'户': %d个)\n", count)
	}
	
	// ==================== 4. 索引操作 ====================
	fmt.Println("\n=== 4. 索引操作测试 ===")
	
	// 创建普通索引
	_, err = db.Exec("CREATE INDEX idx_name ON test_users(name)")
	if err != nil {
		addResult("索引", "创建普通索引", false, err.Error())
		fmt.Printf("  ❌ 创建普通索引失败: %v\n", err)
	} else {
		addResult("索引", "创建普通索引", true, "")
		fmt.Println("  ✅ 创建普通索引成功")
	}
	
	// 创建唯一索引
	_, err = db.Exec("CREATE UNIQUE INDEX idx_email_unique ON test_users(email)")
	if err != nil {
		addResult("索引", "创建唯一索引", false, err.Error())
		fmt.Printf("  ❌ 创建唯一索引失败: %v\n", err)
	} else {
		addResult("索引", "创建唯一索引", true, "")
		fmt.Println("  ✅ 创建唯一索引成功")
	}
	
	// 查看索引
	rows, err = db.Query("SHOW INDEX FROM test_users")
	if err != nil {
		addResult("索引", "查看索引", false, err.Error())
		fmt.Printf("  ❌ 查看索引失败: %v\n", err)
	} else {
		defer rows.Close()
		indexes := make(map[string]bool)
		for rows.Next() {
			var null interface{}
			var table, nonUnique, keyName, seqInIndex, colName string
			var c1, c2, c3, c4, c5, c6, c7, c8, c9, c10 string
			rows.Scan(&table, &nonUnique, &keyName, &seqInIndex, &colName, 
				&c1, &c2, &c3, &c4, &null, &c5, &c6, &c7, &c8, &c9, &c10)
			indexes[keyName] = true
		}
		idxList := []string{}
		for k := range indexes {
			idxList = append(idxList, k)
		}
		addResult("索引", "查看索引", true, fmt.Sprintf("索引:%v", idxList))
		fmt.Printf("  ✅ 查看索引成功: %v\n", idxList)
	}
	
	// ==================== 5. 系统查询 ====================
	fmt.Println("\n=== 5. 系统查询测试 ===")
	
	// 列出用户
	rows, err = db.Query("SELECT user, host FROM mysql.user LIMIT 5")
	if err != nil {
		addResult("系统查询", "列出用户", false, err.Error())
	} else {
		defer rows.Close()
		users := []string{}
		for rows.Next() {
			var user, host string
			rows.Scan(&user, &host)
			users = append(users, fmt.Sprintf("%s@%s", user, host))
		}
		addResult("系统查询", "列出用户", true, fmt.Sprintf("共%d个", len(users)))
		fmt.Printf("  ✅ 列出用户成功 (前%d个)\n", len(users))
	}
	
	// 进程列表
	rows, err = db.Query("SHOW PROCESSLIST")
	if err != nil {
		addResult("系统查询", "进程列表", false, err.Error())
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("系统查询", "进程列表", true, fmt.Sprintf("%d个活跃进程", count))
		fmt.Printf("  ✅ 进程列表成功 (%d个活跃进程)\n", count)
	}
	
	// 服务器状态
	var varName, uptime string
	err = db.QueryRow("SHOW GLOBAL STATUS LIKE 'Uptime'").Scan(&varName, &uptime)
	if err != nil {
		addResult("系统查询", "服务器状态", false, err.Error())
	} else {
		addResult("系统查询", "服务器状态", true, fmt.Sprintf("运行时间:%s秒", uptime))
		fmt.Printf("  ✅ 服务器状态 (运行时间: %s秒)\n", uptime)
	}
	
	// ==================== 6. 错误处理 ====================
	fmt.Println("\n=== 6. 错误处理测试 ===")
	
	// 查询不存在的表
	_, err = db.Query("SELECT * FROM non_existent_table_xxx")
	if err != nil {
		addResult("错误处理", "查询不存在的表", true, "正确报错")
		fmt.Println("  ✅ 查询不存在的表: 正确报错")
	} else {
		addResult("错误处理", "查询不存在的表", false, "应该报错但未报错")
		fmt.Println("  ❌ 查询不存在的表: 应该报错但未报错")
	}
	
	// SQL语法错误
	_, err = db.Query("SELECT * FROM test_users WHERE")
	if err != nil {
		addResult("错误处理", "SQL语法错误", true, "正确报错")
		fmt.Println("  ✅ SQL语法错误: 正确报错")
	} else {
		addResult("错误处理", "SQL语法错误", false, "应该报错但未报错")
		fmt.Println("  ❌ SQL语法错误: 应该报错但未报错")
	}
	
	// 唯一约束冲突
	_, err = db.Exec("INSERT INTO test_users (name, email) VALUES (?, ?)", "测试重复", "a@test.com")
	if err != nil {
		addResult("错误处理", "唯一约束冲突", true, "正确报错")
		fmt.Println("  ✅ 唯一约束冲突: 正确报错")
	} else {
		addResult("错误处理", "唯一约束冲突", false, "应该报错但未报错")
		fmt.Println("  ❌ 唯一约束冲突: 应该报错但未报错")
	}
	
	// 特殊字符
	_, err = db.Exec(
		"INSERT INTO test_users (name, email) VALUES (?, ?)",
		`Special 'quote' "double" \test`, "special@test.com",
	)
	if err != nil {
		addResult("错误处理", "特殊字符处理", false, err.Error())
		fmt.Printf("  ❌ 特殊字符处理失败: %v\n", err)
	} else {
		addResult("错误处理", "特殊字符处理", true, "")
		fmt.Println("  ✅ 特殊字符处理成功")
	}
	
	// ==================== 7. 清理 ====================
	fmt.Println("\n=== 7. 清理 ===")
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))
	if err != nil {
		fmt.Printf("  ⚠️  清理失败: %v\n", err)
	} else {
		fmt.Println("  ✅ 清理成功")
	}
	
	// ==================== 8. 报告 ====================
	generateReport()
}

func generateReport() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    MySQLCTL 测试报告")
	fmt.Println(strings.Repeat("=", 60))
	
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
	
	fmt.Printf("\n📊 汇总: 总测试用例=%d, 通过=%d, 失败=%d, 通过率=%.2f%%\n",
		total, passed, total-passed, float64(passed)*100/float64(total))
	
	fmt.Println("\n📈 按类别统计:")
	for cat, stats := range categoryStats {
		rate := float64(stats.passed) * 100 / float64(stats.total)
		fmt.Printf("  %-10s %d/%d (%.1f%%)\n", cat, stats.passed, stats.total, rate)
	}
	
	failed := []TestResult{}
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}
	
	if len(failed) > 0 {
		fmt.Println("\n❌ 失败的测试:")
		for _, r := range failed {
			fmt.Printf("  [%s] %s → %s\n", r.Category, r.Name, r.Message)
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
}
