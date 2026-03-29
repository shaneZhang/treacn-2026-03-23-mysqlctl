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

func addResult(category, name string, passed bool, message string) {
	results = append(results, TestResult{category, name, passed, message})
}

func main() {
	fmt.Println("=== MySQLCTL 全面测试 ===")

	// ==================== 1. 连接测试 ====================
	fmt.Println("\n=== 1. 连接测试 ===")

	// 1.1 测试正确连接
	err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if err != nil {
		addResult("连接", "使用正确凭据连接", false, err.Error())
		fmt.Println("  ✗ 使用正确凭据连接失败")
	} else {
		addResult("连接", "使用正确凭据连接", true, "")
		fmt.Println("  ✓ 使用正确凭据连接成功")
		db.Disconnect()
	}

	// 1.2 测试错误密码
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "wrongpass", "")
	if err != nil {
		addResult("连接", "使用错误密码连接失败", true, "")
		fmt.Println("  ✓ 使用错误密码正确拒绝")
	} else {
		addResult("连接", "使用错误密码连接失败", false, "应该失败但成功了")
		fmt.Println("  ✗ 错误密码测试失败")
		db.Disconnect()
	}

	// 1.3 测试不可达主机
	fmt.Println("  ? 测试不可达主机(可能超时)...")
	err = db.Connect("192.168.99.99", 3306, "zhangyuqing", "zhangyuqing", "")
	if err != nil {
		addResult("连接", "连接不存在的服务器", true, "")
		fmt.Println("  ✓ 不存在的主机正确拒绝")
	} else {
		addResult("连接", "连接不存在的服务器", false, "应该失败但成功了")
		fmt.Println("  ✗ 不存在主机测试失败")
		db.Disconnect()
	}

	// 1.4 测试指定初始数据库
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysql")
	if err != nil {
		addResult("连接", "连接时指定初始数据库", false, err.Error())
		fmt.Printf("  ✗ 指定初始数据库失败: %v\n", err)
	} else {
		addResult("连接", "连接时指定初始数据库", true, "")
		fmt.Println("  ✓ 连接并指定初始数据库成功")
		db.Disconnect()
	}

	// 1.5 测试空密码
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "", "")
	if err != nil {
		addResult("连接", "使用空密码连接", true, "")
		fmt.Println("  ✓ 空密码正确拒绝")
	} else {
		addResult("连接", "使用空密码连接", false, "应该失败但成功了")
		fmt.Println("  ✗ 空密码测试失败")
		db.Disconnect()
	}

	// 建立持久连接用于后续测试
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if err != nil {
		fmt.Printf("无法建立测试连接: %v\n", err)
		return
	}
	defer db.Disconnect()

	// ==================== 2. 数据库操作测试 ====================
	fmt.Println("\n=== 2. 数据库操作测试 ===")

	// 2.1 列出所有数据库
	rows, err := db.GetDB().Query("SHOW DATABASES")
	if err != nil {
		addResult("数据库", "列出所有数据库", false, err.Error())
		fmt.Printf("  ✗ 列出数据库失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("数据库", "列出所有数据库", true, fmt.Sprintf("共 %d 个数据库", count))
		fmt.Printf("  ✓ 列出所有数据库成功(%d个)\n", count)
	}

	// 2.2 创建测试数据库
	testDB := "mysqlctl_test_db"
	_, err = db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", testDB))
	if err != nil {
		addResult("数据库", "创建新数据库", false, err.Error())
		fmt.Printf("  ✗ 创建数据库失败: %v\n", err)
	} else {
		addResult("数据库", "创建新数据库", true, "")
		fmt.Println("  ✓ 创建测试数据库成功")
	}

	// 2.3 验证数据库存在
	rows, err = db.GetDB().Query("SHOW DATABASES")
	if err == nil {
		defer rows.Close()
		found := false
		for rows.Next() {
			var name string
			rows.Scan(&name)
			if name == testDB {
				found = true
				break
			}
		}
		if found {
			addResult("数据库", "验证新数据库存在", true, "")
			fmt.Println("  ✓ 验证新数据库存在")
		} else {
			addResult("数据库", "验证新数据库存在", false, "数据库未找到")
			fmt.Println("  ✗ 验证数据库存在失败")
		}
	}

	// ==================== 3. 表操作测试 ====================
	fmt.Println("\n=== 3. 表操作测试 ===")

	// 切换到测试数据库
	_, err = db.GetDB().Exec(fmt.Sprintf("USE `%s`", testDB))
	if err != nil {
		fmt.Printf("无法切换到测试数据库: %v\n", err)
	} else {
		fmt.Println("  ✓ 切换到测试数据库")
	}

	// 3.1 创建测试表
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
		fmt.Printf("  ✗ 创建测试表失败: %v\n", err)
	} else {
		addResult("表操作", "创建测试表", true, "")
		fmt.Println("  ✓ 创建测试表成功")
	}

	// 3.2 列出表
	rows, err = db.GetDB().Query("SHOW TABLES")
	if err != nil {
		addResult("表操作", "列出当前数据库表", false, err.Error())
		fmt.Printf("  ✗ 列出表失败: %v\n", err)
	} else {
		defer rows.Close()
		tables := []string{}
		for rows.Next() {
			var name string
			rows.Scan(&name)
			tables = append(tables, name)
		}
		addResult("表操作", "列出当前数据库表", true, fmt.Sprintf("共 %d 个表", len(tables)))
		fmt.Printf("  ✓ 列出表成功(%d个): %v\n", len(tables), tables)
	}

	// 3.3 查看表结构
	rows, err = db.GetDB().Query("DESCRIBE test_users")
	if err != nil {
		addResult("表操作", "查看表结构", false, err.Error())
		fmt.Printf("  ✗ 查看表结构失败: %v\n", err)
	} else {
		defer rows.Close()
		columns := []string{}
		for rows.Next() {
			var field, typ, null, key, dflt, extra string
			rows.Scan(&field, &typ, &null, &key, &dflt, &extra)
			columns = append(columns, field)
		}
		addResult("表操作", "查看表结构", true, fmt.Sprintf("共 %d 列", len(columns)))
		fmt.Printf("  ✓ 查看表结构成功(列数:%d): %v\n", len(columns), columns)
	}

	// ==================== 4. 数据操作测试 ====================
	fmt.Println("\n=== 4. 数据操作测试 ===")

	// 4.1 插入数据
	result, err := db.GetDB().Exec(`
		INSERT INTO test_users (name, email, age) VALUES (?, ?, ?)
	`, "张三", "zhangsan@example.com", 25)
	if err != nil {
		addResult("数据操作", "INSERT插入数据", false, err.Error())
		fmt.Printf("  ✗ 插入数据失败: %v\n", err)
	} else {
		id, _ := result.LastInsertId()
		addResult("数据操作", "INSERT插入数据", true, fmt.Sprintf("插入ID: %d", id))
		fmt.Printf("  ✓ 插入数据成功(ID: %d)\n", id)
	}

	// 4.2 SELECT查询
	rows, err = db.GetDB().Query("SELECT id, name, email FROM test_users WHERE name = ?", "张三")
	if err != nil {
		addResult("数据操作", "SELECT查询数据", false, err.Error())
		fmt.Printf("  ✗ SELECT查询失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			var id int
			var name, email string
			rows.Scan(&id, &name, &email)
			count++
		}
		addResult("数据操作", "SELECT查询数据", true, fmt.Sprintf("返回 %d 行", count))
		fmt.Printf("  ✓ SELECT查询成功(%d行)\n", count)
	}

	// 4.3 UPDATE更新
	result, err = db.GetDB().Exec(`
		UPDATE test_users SET age = ? WHERE name = ?
	`, 26, "张三")
	if err != nil {
		addResult("数据操作", "UPDATE更新数据", false, err.Error())
		fmt.Printf("  ✗ UPDATE更新失败: %v\n", err)
	} else {
		affected, _ := result.RowsAffected()
		addResult("数据操作", "UPDATE更新数据", true, fmt.Sprintf("影响 %d 行", affected))
		fmt.Printf("  ✓ UPDATE更新成功(影响%d行)\n", affected)
	}

	// 4.4 DELETE删除
	result, err = db.GetDB().Exec("DELETE FROM test_users WHERE name = ?", "张三")
	if err != nil {
		addResult("数据操作", "DELETE删除数据", false, err.Error())
		fmt.Printf("  ✗ DELETE删除失败: %v\n", err)
	} else {
		affected, _ := result.RowsAffected()
		addResult("数据操作", "DELETE删除数据", true, fmt.Sprintf("影响 %d 行", affected))
		fmt.Printf("  ✓ DELETE删除成功(影响%d行)\n", affected)
	}

	// ==================== 5. 索引测试 ====================
	fmt.Println("\n=== 5. 索引测试 ===")

	// 5.1 创建索引
	_, err = db.GetDB().Exec("CREATE INDEX idx_name ON test_users(name)")
	if err != nil {
		addResult("索引", "创建普通索引", false, err.Error())
		fmt.Printf("  ✗ 创建索引失败: %v\n", err)
	} else {
		addResult("索引", "创建普通索引", true, "")
		fmt.Println("  ✓ 创建普通索引成功")
	}

	// 5.2 查看索引
	rows, err = db.GetDB().Query("SHOW INDEX FROM test_users")
	if err != nil {
		addResult("索引", "查看表索引", false, err.Error())
		fmt.Printf("  ✗ 查看索引失败: %v\n", err)
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
		addResult("索引", "查看表索引", true, fmt.Sprintf("共 %d 个索引", len(indexes)))
		fmt.Printf("  ✓ 查看索引成功: %v\n", idxList)
	}

	// ==================== 6. 清理测试数据 ====================
	fmt.Println("\n=== 6. 清理测试数据 ===")

	// 删除测试数据库
	_, err = db.GetDB().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))
	if err != nil {
		fmt.Printf("  ! 清理测试数据库失败: %v\n", err)
	} else {
		fmt.Println("  ✓ 清理测试数据库成功")
	}

	// ==================== 7. 生成测试报告 ====================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    测试报告")
	fmt.Println(strings.Repeat("=", 60))

	total := len(results)
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}

	fmt.Printf("\n总测试用例: %d | 通过: %d | 失败: %d | 通过率: %.2f%%\n",
		total, passed, total-passed, float64(passed)*100/float64(total))

	fmt.Println("\n详细结果:")
	fmt.Println(strings.Repeat("-", 60))

	currentCategory := ""
	for _, r := range results {
		if r.Category != currentCategory {
			fmt.Printf("\n[%s]\n", r.Category)
			currentCategory = r.Category
		}
		status := "✓"
		if !r.Passed {
			status = "✗"
		}
		fmt.Printf("  %s %s", status, r.Name)
		if r.Message != "" {
			fmt.Printf(" (%s)", r.Message)
		}
		fmt.Println()
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
}
