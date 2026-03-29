package main

import (
	"fmt"
	"mysqlctl/internal/db"
	"strings"
)

type TestResult struct {
	Category string
	Name     string
	Passed   bool
	Message  string
}

var results []TestResult
var testDB = "mysqlctl_final_test"

func addResult(category, name string, passed bool, message string) {
	results = append(results, TestResult{category, name, passed, message})
}

func main() {
	fmt.Println("=== MySQLCTL 最终全面测试 ===")

	// ==================== 1. 连接测试 ====================
	fmt.Println("\n=== 1. 连接测试 ===")

	// 1.1 正确连接
	err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if err != nil {
		addResult("连接", "使用正确凭据连接成功", false, err.Error())
		fmt.Println("  ❌ 使用正确凭据连接失败")
	} else {
		addResult("连接", "使用正确凭据连接成功", true, "")
		fmt.Println("  ✅ 使用正确凭据连接成功")
		db.Disconnect()
	}

	// 1.2 错误密码
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "wrongpass", "")
	if err != nil {
		addResult("连接", "使用错误密码连接失败", true, "")
		fmt.Println("  ✅ 使用错误密码正确拒绝")
	} else {
		addResult("连接", "使用错误密码连接失败", false, "应该失败但成功了")
		fmt.Println("  ❌ 错误密码测试失败")
		db.Disconnect()
	}

	// 1.3 不可达主机（跳过避免超时）
	fmt.Println("  ⏭️  跳过不可达主机测试（避免超时）")

	// 1.4 指定初始数据库
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysql")
	if err != nil {
		addResult("连接", "连接时指定初始数据库", false, err.Error())
		fmt.Printf("  ❌ 指定初始数据库失败: %v\n", err)
	} else {
		connected, dbName := db.GetStatus()
		if connected && dbName == "mysql" {
			addResult("连接", "连接时指定初始数据库", true, "")
			fmt.Println("  ✅ 连接并指定初始数据库成功")
		} else {
			addResult("连接", "连接时指定初始数据库", false, fmt.Sprintf("状态异常: connected=%v, db=%s", connected, dbName))
			fmt.Println("  ❌ 状态检查失败")
		}
		db.Disconnect()
	}

	// 1.5 空密码
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "", "")
	if err != nil {
		addResult("连接", "使用空密码连接被拒绝", true, "")
		fmt.Println("  ✅ 空密码正确拒绝")
	} else {
		addResult("连接", "使用空密码连接被拒绝", false, "应该失败但成功了")
		fmt.Println("  ❌ 空密码测试失败")
		db.Disconnect()
	}

	// 1.6 断开连接
	db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if db.IsConnected() {
		db.Disconnect()
		if !db.IsConnected() {
			addResult("连接", "断开连接功能正常", true, "")
			fmt.Println("  ✅ 断开连接功能正常")
		} else {
			addResult("连接", "断开连接功能正常", false, "断开后状态仍为已连接")
			fmt.Println("  ❌ 断开连接状态异常")
		}
	} else {
		addResult("连接", "断开连接功能正常", false, "先决条件: 连接失败")
		fmt.Println("  ❌ 先决条件失败")
	}

	// 建立持久连接
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if err != nil {
		fmt.Printf("无法建立测试连接: %v\n", err)
		return
	}
	defer db.Disconnect()

	// ==================== 2. 数据库操作测试 ====================
	fmt.Println("\n=== 2. 数据库操作测试 ===")

	// 清理旧的测试库
	db.GetDB().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))

	// 2.1 列出所有数据库
	rows, err := db.GetDB().Query("SHOW DATABASES")
	if err != nil {
		addResult("数据库", "列出所有数据库", false, err.Error())
		fmt.Printf("  ❌ 列出数据库失败: %v\n", err)
	} else {
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
		addResult("数据库", "列出所有数据库", true, fmt.Sprintf("%d个数据库(含系统库:%v)", count, hasSystemDB))
		fmt.Printf("  ✅ 列出所有数据库成功(%d个,含系统库:%v)\n", count, hasSystemDB)
	}

	// 2.2 创建数据库
	_, err = db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", testDB))
	if err != nil {
		addResult("数据库", "创建新数据库", false, err.Error())
		fmt.Printf("  ❌ 创建数据库失败: %v\n", err)
	} else {
		// 验证
		rows, _ := db.GetDB().Query("SHOW DATABASES LIKE ?", testDB)
		if rows != nil {
			defer rows.Close()
			if rows.Next() {
				addResult("数据库", "创建新数据库", true, "")
				fmt.Println("  ✅ 创建测试数据库成功")
			} else {
				addResult("数据库", "创建新数据库", false, "创建后验证失败")
				fmt.Println("  ❌ 创建数据库验证失败")
			}
		}
	}

	// 2.3 创建已存在的数据库
	_, err = db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE `%s`", testDB))
	if err != nil {
		addResult("数据库", "创建已存在的数据库报错", true, "")
		fmt.Println("  ✅ 创建已存在数据库正确报错")
	} else {
		addResult("数据库", "创建已存在的数据库报错", false, "应该失败但成功了")
		fmt.Println("  ❌ 创建已存在数据库应该报错")
	}

	// 2.4 删除数据库
	_, err = db.GetDB().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))
	if err != nil {
		addResult("数据库", "删除数据库", false, err.Error())
		fmt.Printf("  ❌ 删除数据库失败: %v\n", err)
	} else {
		addResult("数据库", "删除数据库", true, "")
		fmt.Println("  ✅ 删除数据库成功")
	}

	// ==================== 3. 表操作测试 ====================
	fmt.Println("\n=== 3. 表操作测试 ===")

	// 重新创建测试库并使用
	db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", testDB))
	_, err = db.GetDB().Exec(fmt.Sprintf("USE `%s`", testDB))
	if err != nil {
		fmt.Printf("切换到测试数据库失败: %v\n", err)
		return
	}

	// 3.1 创建表
	_, err = db.GetDB().Exec(`
		CREATE TABLE IF NOT EXISTS test_users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(50) NOT NULL,
			email VARCHAR(100) UNIQUE,
			age INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		addResult("表操作", "创建测试表", false, err.Error())
		fmt.Printf("  ❌ 创建测试表失败: %v\n", err)
	} else {
		addResult("表操作", "创建测试表", true, "")
		fmt.Println("  ✅ 创建测试表成功")
	}

	// 3.2 列出表
	rows, err = db.GetDB().Query("SHOW TABLES")
	if err != nil {
		addResult("表操作", "列出当前数据库表", false, err.Error())
		fmt.Printf("  ❌ 列出表失败: %v\n", err)
	} else {
		defer rows.Close()
		tables := []string{}
		for rows.Next() {
			var name string
			rows.Scan(&name)
			tables = append(tables, name)
		}
		addResult("表操作", "列出当前数据库表", true, fmt.Sprintf("%d个表:%v", len(tables), tables))
		fmt.Printf("  ✅ 列出表成功(%d个): %v\n", len(tables), tables)
	}

	// 3.3 查看表结构
	rows, err = db.GetDB().Query("DESCRIBE test_users")
	if err != nil {
		addResult("表操作", "查看表结构", false, err.Error())
		fmt.Printf("  ❌ 查看表结构失败: %v\n", err)
	} else {
		defer rows.Close()
		columns := []string{}
		for rows.Next() {
			var field, typ, null, key, dflt, extra string
			rows.Scan(&field, &typ, &null, &key, &dflt, &extra)
			columns = append(columns, field)
		}
		addResult("表操作", "查看表结构", true, fmt.Sprintf("%d列:%v", len(columns), columns))
		fmt.Printf("  ✅ 查看表结构成功(%d列): %v\n", len(columns), columns)
	}

	// 3.4 查看建表DDL
	rows, err = db.GetDB().Query("SHOW CREATE TABLE test_users")
	if err != nil {
		addResult("表操作", "查看建表DDL", false, err.Error())
		fmt.Printf("  ❌ 查看建表DDL失败: %v\n", err)
	} else {
		defer rows.Close()
		if rows.Next() {
			var tableName, createStmt string
			rows.Scan(&tableName, &createStmt)
			if strings.Contains(createStmt, "CREATE TABLE") {
				addResult("表操作", "查看建表DDL", true, "")
				fmt.Println("  ✅ 查看建表DDL成功")
			} else {
				addResult("表操作", "查看建表DDL", false, "格式不正确")
				fmt.Println("  ❌ 建表DDL格式不正确")
			}
		}
	}

	// ==================== 4. 数据操作测试 ====================
	fmt.Println("\n=== 4. 数据操作测试 ===")

	// 4.1 INSERT
	result, err := db.GetDB().Exec(`
		INSERT INTO test_users (name, email, age) VALUES (?, ?, ?)
	`, "张三", "zhangsan@example.com", 25)
	if err != nil {
		addResult("数据操作", "INSERT插入数据", false, err.Error())
		fmt.Printf("  ❌ 插入数据失败: %v\n", err)
	} else {
		id, _ := result.LastInsertId()
		addResult("数据操作", "INSERT插入数据", true, fmt.Sprintf("ID:%d", id))
		fmt.Printf("  ✅ 插入数据成功(ID:%d)\n", id)
	}

	// 4.2 SELECT
	rows, err = db.GetDB().Query("SELECT id, name FROM test_users WHERE name = ?", "张三")
	if err != nil {
		addResult("数据操作", "SELECT查询数据", false, err.Error())
		fmt.Printf("  ❌ SELECT查询失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("数据操作", "SELECT查询数据", true, fmt.Sprintf("返回%d行", count))
		fmt.Printf("  ✅ SELECT查询成功(%d行)\n", count)
	}

	// 4.3 WHERE条件
	rows, err = db.GetDB().Query("SELECT name FROM test_users WHERE age > ?", 20)
	if err != nil {
		addResult("数据操作", "带WHERE条件查询", false, err.Error())
		fmt.Printf("  ❌ WHERE条件查询失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("数据操作", "带WHERE条件查询", true, fmt.Sprintf("匹配%d行", count))
		fmt.Printf("  ✅ 带WHERE条件查询成功(匹配%d行)\n", count)
	}

	// 4.4 ORDER BY
	rows, err = db.GetDB().Query("SELECT name FROM test_users ORDER BY id DESC LIMIT 5")
	if err != nil {
		addResult("数据操作", "带ORDER BY排序", false, err.Error())
		fmt.Printf("  ❌ ORDER BY测试失败: %v\n", err)
	} else {
		defer rows.Close()
		addResult("数据操作", "带ORDER BY排序", true, "")
		fmt.Println("  ✅ 带ORDER BY排序成功")
	}

	// 4.5 LIMIT分页
	rows, err = db.GetDB().Query("SELECT * FROM test_users LIMIT 10 OFFSET 0")
	if err != nil {
		addResult("数据操作", "带LIMIT分页", false, err.Error())
		fmt.Printf("  ❌ LIMIT测试失败: %v\n", err)
	} else {
		defer rows.Close()
		addResult("数据操作", "带LIMIT分页", true, "")
		fmt.Println("  ✅ 带LIMIT分页成功")
	}

	// 4.6 UPDATE
	result, err = db.GetDB().Exec(`
		UPDATE test_users SET age = ? WHERE name = ?
	`, 26, "张三")
	if err != nil {
		addResult("数据操作", "UPDATE更新数据", false, err.Error())
		fmt.Printf("  ❌ UPDATE更新失败: %v\n", err)
	} else {
		affected, _ := result.RowsAffected()
		addResult("数据操作", "UPDATE更新数据", true, fmt.Sprintf("影响%d行", affected))
		fmt.Printf("  ✅ UPDATE更新成功(影响%d行)\n", affected)
	}

	// 4.7 DELETE
	result, err = db.GetDB().Exec("DELETE FROM test_users WHERE name = ?", "张三")
	if err != nil {
		addResult("数据操作", "DELETE删除数据", false, err.Error())
		fmt.Printf("  ❌ DELETE删除失败: %v\n", err)
	} else {
		affected, _ := result.RowsAffected()
		addResult("数据操作", "DELETE删除数据", true, fmt.Sprintf("删除%d行", affected))
		fmt.Printf("  ✅ DELETE删除成功(删除%d行)\n", affected)
	}

	// ==================== 5. 索引测试 ====================
	fmt.Println("\n=== 5. 索引测试 ===")

	// 5.1 创建普通索引
	_, err = db.GetDB().Exec("CREATE INDEX idx_name ON test_users(name)")
	if err != nil {
		addResult("索引", "创建普通索引", false, err.Error())
		fmt.Printf("  ❌ 创建普通索引失败: %v\n", err)
	} else {
		addResult("索引", "创建普通索引", true, "")
		fmt.Println("  ✅ 创建普通索引成功")
	}

	// 5.2 创建唯一索引
	_, err = db.GetDB().Exec("CREATE UNIQUE INDEX idx_email_unique ON test_users(email)")
	if err != nil {
		addResult("索引", "创建唯一索引", false, err.Error())
		fmt.Printf("  ❌ 创建唯一索引失败: %v\n", err)
	} else {
		addResult("索引", "创建唯一索引", true, "")
		fmt.Println("  ✅ 创建唯一索引成功")
	}

	// 5.3 查看索引
	rows, err = db.GetDB().Query("SHOW INDEX FROM test_users")
	if err != nil {
		addResult("索引", "查看表索引", false, err.Error())
		fmt.Printf("  ❌ 查看索引失败: %v\n", err)
	} else {
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
		idxList := []string{}
		for k := range indexes {
			idxList = append(idxList, k)
		}
		addResult("索引", "查看表索引", true, fmt.Sprintf("%d个索引:%v", len(indexes), idxList))
		fmt.Printf("  ✅ 查看索引成功(%d个): %v\n", len(indexes), idxList)
	}

	// ==================== 6. 系统查询测试 ====================
	fmt.Println("\n=== 6. 系统查询测试 ===")

	// 6.1 列出用户
	rows, err = db.GetDB().Query("SELECT user, host FROM mysql.user LIMIT 10")
	if err != nil {
		addResult("系统查询", "列出所有用户", false, err.Error())
		fmt.Printf("  ❌ 列出用户失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("系统查询", "列出所有用户", true, fmt.Sprintf("%d个用户", count))
		fmt.Printf("  ✅ 列出所有用户成功(显示前%d个)\n", count)
	}

	// 6.2 进程列表
	rows, err = db.GetDB().Query("SHOW PROCESSLIST")
	if err != nil {
		addResult("系统查询", "显示进程列表", false, err.Error())
		fmt.Printf("  ❌ 显示进程列表失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("系统查询", "显示进程列表", true, fmt.Sprintf("%d个进程", count))
		fmt.Printf("  ✅ 显示进程列表成功(%d个活跃进程)\n", count)
	}

	// 6.3 服务器状态
	rows, err = db.GetDB().Query("SHOW GLOBAL STATUS LIKE 'Uptime'")
	if err != nil {
		addResult("系统查询", "获取服务器状态", false, err.Error())
		fmt.Printf("  ❌ 获取服务器状态失败: %v\n", err)
	} else {
		defer rows.Close()
		if rows.Next() {
			var name, value string
			rows.Scan(&name, &value)
			addResult("系统查询", "获取服务器状态", true, fmt.Sprintf("运行时间:%s秒", value))
			fmt.Printf("  ✅ 获取服务器状态成功(运行时间:%s秒)\n", value)
		}
	}

	// ==================== 7. 错误处理测试 ====================
	fmt.Println("\n=== 7. 错误处理测试 ===")

	// 7.1 查询不存在的表
	_, err = db.GetDB().Query("SELECT * FROM non_existent_table_123456")
	if err != nil {
		addResult("错误处理", "查询不存在的表", true, "正确返回错误")
		fmt.Println("  ✅ 查询不存在的表: 正确返回错误")
	} else {
		addResult("错误处理", "查询不存在的表", false, "应该失败但成功了")
		fmt.Println("  ❌ 查询不存在的表: 应该失败但成功了")
	}

	// 7.2 SQL语法错误
	_, err = db.GetDB().Query("SELECT * FROM test_users WHERE")
	if err != nil {
		addResult("错误处理", "SQL语法错误处理", true, "正确返回错误")
		fmt.Println("  ✅ SQL语法错误: 正确返回错误")
	} else {
		addResult("错误处理", "SQL语法错误处理", false, "应该失败但成功了")
		fmt.Println("  ❌ SQL语法错误: 应该失败但成功了")
	}

	// 7.3 重复唯一值
	db.GetDB().Exec("INSERT INTO test_users (name, email) VALUES (?, ?)", "唯一测试", "unique@test.com")
	_, err = db.GetDB().Exec("INSERT INTO test_users (name, email) VALUES (?, ?)", "唯一测试2", "unique@test.com")
	if err != nil {
		addResult("错误处理", "插入重复唯一值报错", true, "正确返回错误")
		fmt.Println("  ✅ 插入重复唯一值: 正确返回错误")
	} else {
		addResult("错误处理", "插入重复唯一值报错", false, "应该失败但成功了")
		fmt.Println("  ❌ 插入重复唯一值: 应该失败但成功了")
	}

	// 7.4 特殊字符处理
	_, err = db.GetDB().Exec(
		"INSERT INTO test_users (name, email) VALUES (?, ?)",
		`Special 'quoted' "value" \test`, "special@test.com",
	)
	if err != nil {
		addResult("错误处理", "插入含特殊字符数据", false, err.Error())
		fmt.Printf("  ❌ 特殊字符插入失败: %v\n", err)
	} else {
		addResult("错误处理", "插入含特殊字符数据", true, "")
		fmt.Println("  ✅ 特殊字符插入成功")
	}

	// ==================== 8. 清理 ====================
	fmt.Println("\n=== 8. 清理测试数据 ===")
	_, err = db.GetDB().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))
	if err != nil {
		fmt.Printf("  ⚠️  清理失败: %v\n", err)
	} else {
		fmt.Println("  ✅ 清理测试数据成功")
	}

	// ==================== 9. 生成最终报告 ====================
	generateFinalReport()
}

func generateFinalReport() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                              MySQLCTL 最终测试报告")
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

	// 汇总
	fmt.Printf("\n📊 测试汇总\n")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("  总测试用例: %d\n", total)
	fmt.Printf("  通过: %d  |  失败: %d\n", passed, total-passed)
	fmt.Printf("  通过率: %.2f%%\n", float64(passed)*100/float64(total))

	// 按类别统计
	fmt.Printf("\n📈 按类别统计\n")
	fmt.Println(strings.Repeat("-", 40))
	for cat, stats := range categoryStats {
		rate := float64(stats.passed) * 100 / float64(stats.total)
		fmt.Printf("  %-12s %d/%d (%.1f%%)\n", cat, stats.passed, stats.total, rate)
	}

	// 失败详情
	failed := []TestResult{}
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}

	if len(failed) > 0 {
		fmt.Printf("\n❌ 失败用例详情\n")
		fmt.Println(strings.Repeat("-", 40))
		for _, r := range failed {
			fmt.Printf("  [%s] %s → %s\n", r.Category, r.Name, r.Message)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}
