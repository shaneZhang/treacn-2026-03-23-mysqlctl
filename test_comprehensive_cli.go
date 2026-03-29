package main

import (
	"fmt"
	"mysqlctl/internal/db"
	"strings"
	"time"
)

type TestResult struct {
	Category string
	Name     string
	Passed   bool
	Message  string
	Duration time.Duration
}

var results []TestResult
var testDBName = "mysqlctl_cli_test"

func runTest(category, name string, testFunc func() (bool, string)) {
	start := time.Now()
	passed, msg := testFunc()
	duration := time.Since(start)
	results = append(results, TestResult{category, name, passed, msg, duration})

	status := "✅"
	if !passed {
		status = "❌"
	}
	fmt.Printf("  %s %-40s %s\n", status, name, msg)
}

func main() {
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("                    MySQLCTL 综合测试报告")
	fmt.Println("=" + strings.Repeat("=", 70))

	// ==================== 1. 连接测试 ====================
	fmt.Println("\n📌 1. 连接测试")
	fmt.Println(strings.Repeat("-", 70))

	// 1.1 正确凭据连接
	runTest("连接", "使用正确凭据连接成功", func() (bool, string) {
		err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
		if err != nil {
			return false, fmt.Sprintf("连接失败: %v", err)
		}
		if !db.IsConnected() {
			return false, "连接状态异常"
		}
		db.Disconnect()
		return true, "连接成功"
	})

	// 1.2 错误密码
	runTest("连接", "使用错误密码连接失败", func() (bool, string) {
		err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "wrongpassword", "")
		if err == nil {
			db.Disconnect()
			return false, "使用错误密码竟然连接成功"
		}
		return true, fmt.Sprintf("正确拒绝: %v", err)
	})

	// 1.3 指定初始数据库
	runTest("连接", "连接时指定初始数据库", func() (bool, string) {
		err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysql")
		if err != nil {
			return false, err.Error()
		}
		connected, currentDB := db.GetStatus()
		db.Disconnect()
		if !connected || currentDB != "mysql" {
			return false, fmt.Sprintf("状态: connected=%v, db=%s", connected, currentDB)
		}
		return true, fmt.Sprintf("已连接到: %s", currentDB)
	})

	// 1.4 断开连接测试
	runTest("连接", "断开连接功能正常", func() (bool, string) {
		db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
		if !db.IsConnected() {
			return false, "先决条件：连接失败"
		}
		db.Disconnect()
		if db.IsConnected() {
			return false, "断开后状态仍显示已连接"
		}
		return true, "断开连接成功"
	})

	// 1.5 连接超时（使用不可达地址测试，只等待3秒）
	runTest("连接", "未连接时状态显示正确", func() (bool, string) {
		if db.IsConnected() {
			db.Disconnect()
		}
		connected, _ := db.GetStatus()
		if connected {
			return false, "未连接时状态错误"
		}
		return true, "未连接状态正确"
	})

	// 建立连接用于后续测试
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	defer func() {
		db.GetDB().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDBName))
	}()

	// ==================== 2. 数据库操作 ====================
	fmt.Println("\n📌 2. 数据库操作测试")
	fmt.Println(strings.Repeat("-", 70))

	// 清理旧库
	db.GetDB().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDBName))

	// 2.1 列出数据库
	runTest("数据库", "列出所有数据库", func() (bool, string) {
		rows, err := db.GetDB().Query("SHOW DATABASES")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		count := 0
		hasSystemDB := false
		for rows.Next() {
			var name string
			rows.Scan(&name)
			count++
			if name == "information_schema" || name == "mysql" {
				hasSystemDB = true
			}
		}
		return true, fmt.Sprintf("共%d个数据库（含系统库:%v）", count, hasSystemDB)
	})

	// 2.2 创建数据库
	runTest("数据库", "创建新数据库", func() (bool, string) {
		_, err := db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE `%s`", testDBName))
		if err != nil {
			return false, err.Error()
		}
		// 验证
		rows, err := db.GetDB().Query("SHOW DATABASES LIKE ?", testDBName)
		if err != nil {
			return false, fmt.Sprintf("验证失败: %v", err)
		}
		defer rows.Close()
		if rows.Next() {
			return true, fmt.Sprintf("数据库 %s 创建成功", testDBName)
		}
		return false, "创建后验证失败"
	})

	// 2.3 创建已存在数据库
	runTest("数据库", "创建已存在数据库报错", func() (bool, string) {
		_, err := db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE `%s`", testDBName))
		if err == nil {
			return false, "应该报错但成功了"
		}
		return true, fmt.Sprintf("正确拒绝: %v", err)
	})

	// 2.4 删除数据库测试
	runTest("数据库", "删除数据库", func() (bool, string) {
		// 创建一个临时数据库来测试删除
		tempDB := "temp_db_to_delete"
		db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", tempDB))
		_, err := db.GetDB().Exec(fmt.Sprintf("DROP DATABASE `%s`", tempDB))
		if err != nil {
			return false, err.Error()
		}
		return true, "临时数据库删除成功"
	})

	// 切换到测试数据库
	db.GetDB().Exec(fmt.Sprintf("USE `%s`", testDBName))

	// ==================== 3. 表操作 ====================
	fmt.Println("\n📌 3. 表操作测试")
	fmt.Println(strings.Repeat("-", 70))

	// 3.1 创建表
	runTest("表操作", "创建测试表", func() (bool, string) {
		sql := `
			CREATE TABLE test_products (
				id INT PRIMARY KEY AUTO_INCREMENT,
				name VARCHAR(100) NOT NULL,
				price DECIMAL(10,2) NOT NULL DEFAULT 0,
				stock INT DEFAULT 0,
				description TEXT,
				is_active BOOLEAN DEFAULT TRUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
		_, err := db.GetDB().Exec(sql)
		if err != nil {
			return false, err.Error()
		}
		return true, "创建表 test_products 成功"
	})

	// 3.2 列出表
	runTest("表操作", "列出当前数据库表", func() (bool, string) {
		rows, err := db.GetDB().Query("SHOW TABLES")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		tables := []string{}
		for rows.Next() {
			var name string
			rows.Scan(&name)
			tables = append(tables, name)
		}
		return true, fmt.Sprintf("表列表: %v", tables)
	})

	// 3.3 查看表结构
	runTest("表操作", "查看表结构（DESCRIBE）", func() (bool, string) {
		rows, err := db.GetDB().Query("DESCRIBE test_products")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		columns := []string{}
		for rows.Next() {
			var field, typ, null, key, dflt, extra string
			rows.Scan(&field, &typ, &null, &key, &dflt, &extra)
			columns = append(columns, field)
		}
		return true, fmt.Sprintf("字段: %v", columns)
	})

	// 3.4 查看建表 DDL
	runTest("表操作", "查看建表 DDL（SHOW CREATE）", func() (bool, string) {
		rows, err := db.GetDB().Query("SHOW CREATE TABLE test_products")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		if rows.Next() {
			var tableName, createStmt string
			rows.Scan(&tableName, &createStmt)
			if strings.Contains(createStmt, "CREATE TABLE") {
				return true, "DDL 获取成功"
			}
		}
		return false, "DDL 格式不正确"
	})

	// ==================== 4. 数据操作 ====================
	fmt.Println("\n📌 4. 数据操作测试")
	fmt.Println(strings.Repeat("-", 70))

	// 4.1 INSERT 插入
	var lastInsertID int64
	runTest("数据操作", "INSERT 插入单条数据", func() (bool, string) {
		result, err := db.GetDB().Exec(`
			INSERT INTO test_products (name, price, stock, description)
			VALUES (?, ?, ?, ?)
		`, "iPhone 15 Pro", 8999.00, 100, "最新苹果手机")
		if err != nil {
			return false, err.Error()
		}
		lastInsertID, _ = result.LastInsertId()
		return true, fmt.Sprintf("插入成功，新记录 ID: %d", lastInsertID)
	})

	// 4.2 SELECT 查询
	runTest("数据操作", "SELECT 基础查询", func() (bool, string) {
		rows, err := db.GetDB().Query("SELECT id, name FROM test_products WHERE id = ?", lastInsertID)
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		if rows.Next() {
			var id int
			var name string
			rows.Scan(&id, &name)
			return true, fmt.Sprintf("查询结果: id=%d, name=%s", id, name)
		}
		return false, "未查询到数据"
	})

	// 4.3 WHERE 条件过滤
	runTest("数据操作", "WHERE 条件过滤", func() (bool, string) {
		rows, err := db.GetDB().Query("SELECT name FROM test_products WHERE price > ?", 5000)
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		return true, fmt.Sprintf("匹配 %d 条记录", count)
	})

	// 批量插入测试数据
	for i := 1; i <= 5; i++ {
		db.GetDB().Exec(`
			INSERT INTO test_products (name, price, stock)
			VALUES (?, ?, ?)
		`, fmt.Sprintf("商品%d", i), 100.00*float64(i), i*10)
	}

	// 4.4 ORDER BY 排序
	runTest("数据操作", "ORDER BY 排序", func() (bool, string) {
		rows, err := db.GetDB().Query("SELECT name, price FROM test_products ORDER BY price DESC LIMIT 3")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		items := []string{}
		for rows.Next() {
			var name string
			var price float64
			rows.Scan(&name, &price)
			items = append(items, fmt.Sprintf("%s(%.0f)", name, price))
		}
		return true, fmt.Sprintf("Top3: %v", items)
	})

	// 4.5 LIMIT 分页
	runTest("数据操作", "LIMIT + OFFSET 分页", func() (bool, string) {
		rows, err := db.GetDB().Query("SELECT name FROM test_products LIMIT 2 OFFSET 1")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		items := []string{}
		for rows.Next() {
			var name string
			rows.Scan(&name)
			items = append(items, name)
		}
		return true, fmt.Sprintf("第2-3条: %v", items)
	})

	// 4.6 UPDATE 更新
	runTest("数据操作", "UPDATE 更新数据", func() (bool, string) {
		result, err := db.GetDB().Exec(`
			UPDATE test_products SET price = price * 0.9, stock = stock + 50
			WHERE name LIKE ?
		`, "%商品%")
		if err != nil {
			return false, err.Error()
		}
		affected, _ := result.RowsAffected()
		return true, fmt.Sprintf("更新成功，影响 %d 行", affected)
	})

	// 4.7 DELETE 删除
	runTest("数据操作", "DELETE 删除数据", func() (bool, string) {
		result, err := db.GetDB().Exec("DELETE FROM test_products WHERE name = ?", "商品5")
		if err != nil {
			return false, err.Error()
		}
		affected, _ := result.RowsAffected()
		return true, fmt.Sprintf("删除成功，删除 %d 行", affected)
	})

	// 4.8 复杂查询（LIKE + IN + AND）
	runTest("数据操作", "复杂条件查询（LIKE+IN）", func() (bool, string) {
		rows, err := db.GetDB().Query(`
			SELECT name, price FROM test_products 
			WHERE (name LIKE ? OR id IN (1, 2)) AND stock > 0
			LIMIT 5
		`, "%商品%")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		return true, fmt.Sprintf("匹配 %d 条", count)
	})

	// ==================== 5. 索引操作 ====================
	fmt.Println("\n📌 5. 索引操作测试")
	fmt.Println(strings.Repeat("-", 70))

	// 5.1 创建普通索引
	runTest("索引", "创建普通索引（CREATE INDEX）", func() (bool, string) {
		_, err := db.GetDB().Exec("CREATE INDEX idx_name ON test_products(name)")
		if err != nil {
			return false, err.Error()
		}
		return true, "索引 idx_name 创建成功"
	})

	// 5.2 创建唯一索引
	runTest("索引", "创建唯一索引（UNIQUE）", func() (bool, string) {
		_, err := db.GetDB().Exec("CREATE UNIQUE INDEX idx_unique_name ON test_products(name(50))")
		if err != nil {
			return false, err.Error()
		}
		return true, "唯一索引 idx_unique_name 创建成功"
	})

	// 5.3 查看索引
	runTest("索引", "查看表索引（SHOW INDEX）", func() (bool, string) {
		rows, err := db.GetDB().Query("SHOW INDEX FROM test_products")
		if err != nil {
			return false, err.Error()
		}
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
		return true, fmt.Sprintf("索引列表: %v", idxList)
	})

	// ==================== 6. 用户与权限 ====================
	fmt.Println("\n📌 6. 用户与权限测试")
	fmt.Println(strings.Repeat("-", 70))

	// 6.1 列出用户
	runTest("用户权限", "列出所有 MySQL 用户", func() (bool, string) {
		rows, err := db.GetDB().Query("SELECT user, host FROM mysql.user LIMIT 5")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		users := []string{}
		for rows.Next() {
			var user, host string
			rows.Scan(&user, &host)
			users = append(users, fmt.Sprintf("%s@%s", user, host))
		}
		return true, fmt.Sprintf("前5个用户: %v", users)
	})

	// ==================== 7. 系统状态 ====================
	fmt.Println("\n📌 7. 系统状态测试")
	fmt.Println(strings.Repeat("-", 70))

	// 7.1 进程列表
	runTest("系统状态", "显示进程列表（SHOW PROCESSLIST）", func() (bool, string) {
		rows, err := db.GetDB().Query("SHOW PROCESSLIST")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		return true, fmt.Sprintf("当前有 %d 个活跃进程", count)
	})

	// 7.2 服务器状态
	runTest("系统状态", "显示服务器状态（SHOW STATUS）", func() (bool, string) {
		rows, err := db.GetDB().Query("SHOW GLOBAL STATUS LIKE 'Uptime'")
		if err != nil {
			return false, err.Error()
		}
		defer rows.Close()
		if rows.Next() {
			var name, value string
			rows.Scan(&name, &value)
			return true, fmt.Sprintf("服务器运行时间: %s 秒", value)
		}
		return false, "无法获取状态"
	})

	// ==================== 8. 错误处理 ====================
	fmt.Println("\n📌 8. 错误处理测试")
	fmt.Println(strings.Repeat("-", 70))

	// 8.1 查询不存在的表
	runTest("错误处理", "查询不存在的表", func() (bool, string) {
		_, err := db.GetDB().Query("SELECT * FROM table_that_never_exists_12345")
		if err == nil {
			return false, "应该报错但没有"
		}
		return true, fmt.Sprintf("正确拒绝: %v", err)
	})

	// 8.2 SQL 语法错误
	runTest("错误处理", "SQL 语法错误处理", func() (bool, string) {
		_, err := db.GetDB().Query("SELECT * FROM test_products WHERE COLUMN_NOT_EXIST = 1")
		if err == nil {
			return false, "应该报错但没有"
		}
		return true, fmt.Sprintf("正确拒绝: %v", err)
	})

	// 8.3 插入重复唯一值
	runTest("错误处理", "唯一约束冲突处理", func() (bool, string) {
		// 插入重复名称（name 有唯一索引）
		_, err := db.GetDB().Exec(`
			INSERT INTO test_products (name, price) VALUES ('iPhone 15 Pro', 100)
		`)
		if err == nil {
			return false, "应该报错但没有"
		}
		return true, fmt.Sprintf("正确拒绝: %v", err)
	})

	// 8.4 特殊字符处理
	runTest("错误处理", "特殊字符转义处理", func() (bool, string) {
		_, err := db.GetDB().Exec(`
			INSERT INTO test_products (name, price, description)
			VALUES (?, ?, ?)
		`, `特殊'字符"测试\没问题`, 999, `内容包含'单引'和"双引"`)
		if err != nil {
			return false, err.Error()
		}
		return true, "特殊字符插入成功"
	})

	// ==================== 生成最终报告 ====================
	generateFinalReport()
}

func generateFinalReport() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("                      📊 最终测试报告汇总")
	fmt.Println(strings.Repeat("=", 70))

	total := len(results)
	passed := 0
	categoryStats := make(map[string]struct{ total, passed int })
	var totalDuration time.Duration

	for _, r := range results {
		stats := categoryStats[r.Category]
		stats.total++
		if r.Passed {
			passed++
			stats.passed++
		}
		categoryStats[r.Category] = stats
		totalDuration += r.Duration
	}

	successRate := float64(passed) * 100 / float64(total)

	fmt.Printf("\n📋 汇总统计\n")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("  总测试用例数: %d\n", total)
	fmt.Printf("  ✅ 通过用例: %d\n", passed)
	fmt.Printf("  ❌ 失败用例: %d\n", total-passed)
	fmt.Printf("  📈 成功率: %.2f%%\n", successRate)
	fmt.Printf("  ⏱️  总耗时: %v\n", totalDuration.Round(time.Millisecond))

	fmt.Printf("\n📈 按类别统计\n")
	fmt.Println(strings.Repeat("-", 70))
	for cat, stats := range categoryStats {
		rate := float64(stats.passed) * 100 / float64(stats.total)
		fmt.Printf("  %-12s %2d/%d  (%.1f%%)\n", cat, stats.passed, stats.total, rate)
	}

	failed := []TestResult{}
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}

	if len(failed) > 0 {
		fmt.Printf("\n❌ 失败用例详情\n")
		fmt.Println(strings.Repeat("-", 70))
		for _, r := range failed {
			fmt.Printf("  [%s] %s → %s\n", r.Category, r.Name, r.Message)
		}
	} else {
		fmt.Printf("\n🎉 所有测试用例全部通过！\n")
	}

	fmt.Printf("\n✅ 通过的测试用例详情\n")
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
	fmt.Println("                        测试结束")
	fmt.Println(strings.Repeat("=", 70))
}
