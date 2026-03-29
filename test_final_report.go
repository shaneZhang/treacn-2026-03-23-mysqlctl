package main

import (
	"fmt"
	"strings"
	"mysqlctl/internal/db"
)

type TestCase struct {
	Category string
	Name     string
	TestFunc func() (bool, string)
}

type TestResult struct {
	Category string
	Name     string
	Passed   bool
	Message  string
}

var results []TestResult

func runTest(tc TestCase) {
	passed, msg := tc.TestFunc()
	results = append(results, TestResult{tc.Category, tc.Name, passed, msg})
}

func main() {
	fmt.Println("=== MySQLCTL 最终全面测试 ===")
	
	// 测试用例列表
	testCases := []TestCase{
		// ==================== 连接测试 ====================
		{"连接", "使用正确凭据连接成功", testConnectSuccess},
		{"连接", "使用错误密码连接失败", testWrongPassword},
		{"连接", "连接不存在的服务器", testUnreachableHost},
		{"连接", "连接时指定初始数据库", testConnectWithDatabase},
		{"连接", "使用空密码连接被拒绝", testEmptyPassword},
		{"连接", "断开连接功能正常", testDisconnect},
		
		// ==================== 数据库操作测试 ====================
		{"数据库", "列出所有数据库", testListDatabases},
		{"数据库", "创建新数据库", testCreateDatabase},
		{"数据库", "创建已存在的数据库", testCreateExistingDatabase},
		{"数据库", "删除数据库", testDropDatabase},
		{"数据库", "删除不存在的数据库", testDropNonExistingDB},
		
		// ==================== 表操作测试 ====================
		{"表操作", "创建测试表", testCreateTable},
		{"表操作", "列出当前数据库表", testListTables},
		{"表操作", "查看表结构", testDescribeTable},
		{"表操作", "查看建表DDL", testShowCreateTable},
		
		// ==================== 数据操作测试 ====================
		{"数据操作", "INSERT插入数据", testInsertData},
		{"数据操作", "SELECT查询数据", testSelectData},
		{"数据操作", "带WHERE条件查询", testSelectWithWhere},
		{"数据操作", "带ORDER BY排序", testSelectWithOrderBy},
		{"数据操作", "带LIMIT分页", testSelectWithLimit},
		{"数据操作", "UPDATE更新数据", testUpdateData},
		{"数据操作", "DELETE删除数据", testDeleteData},
		
		// ==================== 索引测试 ====================
		{"索引", "创建普通索引", testCreateIndex},
		{"索引", "创建唯一索引", testCreateUniqueIndex},
		{"索引", "查看表索引", testShowIndex},
		
		// ==================== 系统查询测试 ====================
		{"系统查询", "列出所有用户", testListUsers},
		{"系统查询", "显示进程列表", testProcessList},
		{"系统查询", "获取服务器状态", testShowStatus},
		
		// ==================== 错误处理测试 ====================
		{"错误处理", "查询不存在的表", testQueryNonExistTable},
		{"错误处理", "SQL语法错误处理", testSQLSyntaxError},
		{"错误处理", "插入重复唯一值", testDuplicateEntry},
	}
	
	// 运行所有测试
	for _, tc := range testCases {
		fmt.Printf("测试: %s - %s...\n", tc.Category, tc.Name)
		runTest(tc)
	}
	
	// 生成报告
	generateReport()
	
	// 清理测试数据
	cleanup()
}

// ==================== 测试函数 ====================

func testConnectSuccess() (bool, string) {
	err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if err != nil {
		return false, fmt.Sprintf("连接失败: %v", err)
	}
	db.Disconnect()
	return true, ""
}

func testWrongPassword() (bool, string) {
	err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "wrongpass", "")
	if err == nil {
		db.Disconnect()
		return false, "使用错误密码应该连接失败"
	}
	return true, fmt.Sprintf("正确拒绝: %v", err)
}

func testUnreachableHost() (bool, string) {
	err := db.Connect("192.168.99.99", 3306, "zhangyuqing", "zhangyuqing", "")
	if err == nil {
		db.Disconnect()
		return false, "连接不存在的服务器应该失败"
	}
	return true, ""
}

func testConnectWithDatabase() (bool, string) {
	err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysql")
	if err != nil {
		return false, fmt.Sprintf("连接失败: %v", err)
	}
	connected, dbName := db.GetStatus()
	db.Disconnect()
	if !connected || dbName != "mysql" {
		return false, fmt.Sprintf("状态不正确 - connected:%v, dbName:%s", connected, dbName)
	}
	return true, ""
}

func testEmptyPassword() (bool, string) {
	err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "", "")
	if err == nil {
		db.Disconnect()
		return false, "使用空密码应该连接失败"
	}
	return true, ""
}

func testDisconnect() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if !db.IsConnected() {
		return false, "连接后状态应为已连接"
	}
	err := db.Disconnect()
	if err != nil {
		return false, fmt.Sprintf("断开连接失败: %v", err)
	}
	if db.IsConnected() {
		return false, "断开连接后状态应为未连接"
	}
	return true, ""
}

func testListDatabases() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SHOW DATABASES")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
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
	return true, fmt.Sprintf("共%d个数据库, 包含系统库:%v", count, hasSystemDB)
}

func testCreateDatabase() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	testDB := "mysqlctl_test_create"
	_, err := db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", testDB))
	if err != nil {
		return false, fmt.Sprintf("创建失败: %v", err)
	}
	
	// 验证是否存在
	rows, _ := db.GetDB().Query("SHOW DATABASES LIKE ?", testDB)
	defer rows.Close()
	if !rows.Next() {
		return false, "创建的数据库不存在"
	}
	
	return true, ""
}

func testCreateExistingDatabase() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	testDB := "mysqlctl_test_exist"
	db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", testDB))
	
	// 再次创建应该报错（如果不加 IF NOT EXISTS）
	_, err := db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE `%s`", testDB))
	if err == nil {
		return false, "创建已存在的数据库应该报错"
	}
	
	return true, ""
}

func testDropDatabase() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	testDB := "mysqlctl_test_drop"
	db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", testDB))
	
	_, err := db.GetDB().Exec(fmt.Sprintf("DROP DATABASE `%s`", testDB))
	if err != nil {
		return false, fmt.Sprintf("删除失败: %v", err)
	}
	
	return true, ""
}

func testDropNonExistingDB() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	_, err := db.GetDB().Exec("DROP DATABASE `non_existing_db_12345`")
	if err == nil {
		return false, "删除不存在的数据库应该报错"
	}
	
	return true, ""
}

func testCreateTable() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	testDB := "mysqlctl_test_table"
	db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", testDB))
	db.GetDB().Exec(fmt.Sprintf("USE `%s`", testDB))
	
	_, err := db.GetDB().Exec(`
		CREATE TABLE test_table (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(50) NOT NULL,
			email VARCHAR(100) UNIQUE
		)
	`)
	if err != nil {
		return false, fmt.Sprintf("创建表失败: %v", err)
	}
	
	return true, ""
}

func testListTables() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SHOW TABLES")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	tables := []string{}
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}
	
	return true, fmt.Sprintf("表列表: %v", tables)
}

func testDescribeTable() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("DESCRIBE test_table")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	columns := []string{}
	for rows.Next() {
		var field, typ, null, key, dflt, extra string
		rows.Scan(&field, &typ, &null, &key, &dflt, &extra)
		columns = append(columns, field)
	}
	
	return true, fmt.Sprintf("列: %v", columns)
}

func testShowCreateTable() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SHOW CREATE TABLE test_table")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	if rows.Next() {
		var tableName, createStmt string
		err := rows.Scan(&tableName, &createStmt)
		if err != nil {
			return false, fmt.Sprintf("读取结果失败: %v", err)
		}
		if !strings.Contains(createStmt, "CREATE TABLE") {
			return false, "建表语句格式不正确"
		}
	}
	
	return true, ""
}

func testInsertData() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	result, err := db.GetDB().Exec(`
		INSERT INTO test_table (name, email) VALUES (?, ?)
	`, "测试用户", "test@example.com")
	if err != nil {
		return false, fmt.Sprintf("插入失败: %v", err)
	}
	
	id, _ := result.LastInsertId()
	return true, fmt.Sprintf("插入ID: %d", id)
}

func testSelectData() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SELECT * FROM test_table")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		count++
	}
	
	return true, fmt.Sprintf("返回%d行", count)
}

func testSelectWithWhere() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SELECT * FROM test_table WHERE name = ?", "测试用户")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		count++
	}
	
	return true, fmt.Sprintf("匹配%d行", count)
}

func testSelectWithOrderBy() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SELECT * FROM test_table ORDER BY id DESC")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	return true, "排序查询成功"
}

func testSelectWithLimit() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SELECT * FROM test_table LIMIT 10 OFFSET 0")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	return true, "分页查询成功"
}

func testUpdateData() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	result, err := db.GetDB().Exec(`
		UPDATE test_table SET email = ? WHERE name = ?
	`, "updated@example.com", "测试用户")
	if err != nil {
		return false, fmt.Sprintf("更新失败: %v", err)
	}
	
	affected, _ := result.RowsAffected()
	return true, fmt.Sprintf("影响%d行", affected)
}

func testDeleteData() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	result, err := db.GetDB().Exec("DELETE FROM test_table WHERE name = ?", "测试用户")
	if err != nil {
		return false, fmt.Sprintf("删除失败: %v", err)
	}
	
	affected, _ := result.RowsAffected()
	return true, fmt.Sprintf("删除%d行", affected)
}

func testCreateIndex() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	_, err := db.GetDB().Exec("CREATE INDEX idx_name ON test_table(name)")
	if err != nil {
		return false, fmt.Sprintf("创建索引失败: %v", err)
	}
	
	return true, ""
}

func testCreateUniqueIndex() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	_, err := db.GetDB().Exec("CREATE UNIQUE INDEX idx_email_unique ON test_table(email)")
	if err != nil {
		return false, fmt.Sprintf("创建唯一索引失败: %v", err)
	}
	
	return true, ""
}

func testShowIndex() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SHOW INDEX FROM test_table")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	indexes := make(map[string]bool)
	for rows.Next() {
		var null interface{}
		var table, nonUnique, keyName, seqInIndex, colName string
		var collation, cardinality, subPart, packed, indexType, comment, indexComment string
		var visible, expression string
		rows.Scan(&table, &nonUnique, &keyName, &seqInIndex, &colName, &collation, 
			&cardinality, &subPart, &packed, &null, &indexType, &comment, 
			&indexComment, &visible, &expression)
		indexes[keyName] = true
	}
	
	if len(indexes) == 0 {
		return false, "未找到任何索引"
	}
	
	idxList := []string{}
	for k := range indexes {
		idxList = append(idxList, k)
	}
	return true, fmt.Sprintf("索引: %v", idxList)
}

func testListUsers() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SELECT user, host FROM mysql.user")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		count++
	}
	
	return true, fmt.Sprintf("共%d个用户", count)
}

func testProcessList() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SHOW PROCESSLIST")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		count++
	}
	
	return true, fmt.Sprintf("共%d个活跃进程", count)
}

func testShowStatus() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	rows, err := db.GetDB().Query("SHOW GLOBAL STATUS LIKE 'Uptime'")
	if err != nil {
		return false, fmt.Sprintf("查询失败: %v", err)
	}
	defer rows.Close()
	
	if rows.Next() {
		var name, value string
		rows.Scan(&name, &value)
		return true, fmt.Sprintf("运行时间: %s秒", value)
	}
	
	return false, "未获取到状态信息"
}

func testQueryNonExistTable() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	_, err := db.GetDB().Query("SELECT * FROM non_existent_table_12345")
	if err == nil {
		return false, "查询不存在的表应该报错"
	}
	
	return true, "正确处理表不存在的错误"
}

func testSQLSyntaxError() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	_, err := db.GetDB().Query("SELECT * FROM test_table WHERE") // 语法错误
	if err == nil {
		return false, "SQL语法错误应该被捕获"
	}
	
	return true, "正确处理SQL语法错误"
}

func testDuplicateEntry() (bool, string) {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysqlctl_test_table")
	defer db.Disconnect()
	
	// 先插入一条数据
	db.GetDB().Exec("INSERT INTO test_table (name, email) VALUES (?, ?)", "重复测试", "duplicate@example.com")
	
	// 再插入相同email（email有唯一索引）
	_, err := db.GetDB().Exec("INSERT INTO test_table (name, email) VALUES (?, ?)", "重复测试2", "duplicate@example.com")
	if err == nil {
		return false, "重复唯一值应该报错"
	}
	
	return true, "正确处理唯一约束冲突"
}

// ==================== 辅助函数 ====================

func generateReport() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                              MySQLCTL 测试报告")
	fmt.Println(strings.Repeat("=", 80))
	
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
	
	// 汇总统计
	fmt.Printf("\n📊 测试汇总统计\n")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("  总测试用例: %d\n", total)
	fmt.Printf("  通过: %d  |  失败: %d\n", passed, total-passed)
	fmt.Printf("  通过率: %.2f%%\n", float64(passed)*100/float64(total))
	
	// 按分类统计
	fmt.Printf("\n📈 按类别统计\n")
	fmt.Println(strings.Repeat("-", 80))
	for cat, stats := range categoryStats {
		rate := float64(stats.passed) * 100 / float64(stats.total)
		fmt.Printf("  %-15s 通过: %d/%d (%.1f%%)\n", cat, stats.passed, stats.total, rate)
	}
	
	// 详细结果
	fmt.Printf("\n🔍 详细测试结果\n")
	fmt.Println(strings.Repeat("-", 80))
	
	currentCategory := ""
	for _, r := range results {
		if r.Category != currentCategory {
			fmt.Printf("\n  [%s]\n", r.Category)
			currentCategory = r.Category
		}
		status := "✅"
		if !r.Passed {
			status = "❌"
		}
		fmt.Printf("    %s %s", status, r.Name)
		if r.Message != "" {
			fmt.Printf(" → %s", r.Message)
		}
		fmt.Println()
	}
	
	// 失败用例详情
	failed := []TestResult{}
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}
	
	if len(failed) > 0 {
		fmt.Printf("\n❌ 失败的测试用例\n")
		fmt.Println(strings.Repeat("-", 80))
		for _, r := range failed {
			fmt.Printf("  [%s] %s → %s\n", r.Category, r.Name, r.Message)
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 80))
}

func cleanup() {
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	defer db.Disconnect()
	
	// 删除所有测试数据库
	testDBs := []string{
		"mysqlctl_test_create",
		"mysqlctl_test_exist",
		"mysqlctl_test_table",
		"mysqlctl_test_db",
		"mysqlctl_test_cli",
	}
	
	for _, dbName := range testDBs {
		db.GetDB().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
	}
	
	fmt.Println("\n🧹 测试数据清理完成")
}
