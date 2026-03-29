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
var testDB = "mysqlctl_test_cli"

func addResult(category, name string, passed bool, message string) {
	results = append(results, TestResult{category, name, passed, message})
}

func main() {
	fmt.Println("=== MySQLCTL CLI 高级功能测试 ===")

	// 建立连接
	err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if err != nil {
		fmt.Printf("无法建立测试连接: %v\n", err)
		return
	}
	defer db.Disconnect()

	// ==================== 1. 用户管理测试 ====================
	fmt.Println("\n=== 1. 用户管理测试 ===")

	// 1.1 列出所有用户
	rows, err := db.GetDB().Query("SELECT user, host FROM mysql.user")
	if err != nil {
		addResult("用户管理", "列出所有用户", false, err.Error())
		fmt.Printf("  ✗ 列出用户失败: %v\n", err)
	} else {
		defer rows.Close()
		users := []string{}
		for rows.Next() {
			var user, host string
			rows.Scan(&user, &host)
			users = append(users, fmt.Sprintf("%s@%s", user, host))
		}
		addResult("用户管理", "列出所有用户", true, fmt.Sprintf("共 %d 个用户", len(users)))
		fmt.Printf("  ✓ 列出所有用户成功(%d个)\n", len(users))
	}

	// ==================== 2. 进程列表测试 ====================
	fmt.Println("\n=== 2. 进程列表测试 ===")

	rows, err = db.GetDB().Query("SHOW PROCESSLIST")
	if err != nil {
		addResult("系统状态", "显示进程列表", false, err.Error())
		fmt.Printf("  ✗ 显示进程列表失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("系统状态", "显示进程列表", true, fmt.Sprintf("共 %d 个进程", count))
		fmt.Printf("  ✓ 显示进程列表成功(%d个活跃进程)\n", count)
	}

	// ==================== 3. 服务器状态测试 ====================
	fmt.Println("\n=== 3. 服务器状态测试 ===")

	rows, err = db.GetDB().Query("SHOW GLOBAL STATUS LIKE 'Uptime'")
	if err != nil {
		addResult("系统状态", "获取服务器状态", false, err.Error())
		fmt.Printf("  ✗ 获取服务器状态失败: %v\n", err)
	} else {
		defer rows.Close()
		if rows.Next() {
			var name, value string
			rows.Scan(&name, &value)
			addResult("系统状态", "获取服务器状态", true, fmt.Sprintf("运行时间: %s秒", value))
			fmt.Printf("  ✓ 获取服务器状态成功(运行时间: %s秒)\n", value)
		}
	}

	// ==================== 4. 高级查询测试 ====================
	fmt.Println("\n=== 4. 高级查询测试 ===")

	// 创建测试数据库和表
	_, err = db.GetDB().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", testDB))
	if err != nil {
		fmt.Printf("创建测试数据库失败: %v\n", err)
		return
	}
	_, err = db.GetDB().Exec(fmt.Sprintf("USE `%s`", testDB))
	if err != nil {
		fmt.Printf("切换到测试数据库失败: %v\n", err)
		return
	}

	// 创建测试表并插入数据
	_, err = db.GetDB().Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			category VARCHAR(50),
			stock INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Printf("创建测试表失败: %v\n", err)
		return
	}

	// 批量插入数据
	products := []struct {
		name, category string
		price          float64
		stock          int
	}{
		{"iPhone 15", "手机", 7999.00, 100},
		{"MacBook Pro", "电脑", 14999.00, 50},
		{"AirPods Pro", "配件", 1899.00, 200},
		{"iPad Air", "平板", 4799.00, 80},
		{"Apple Watch", "手表", 2999.00, 120},
	}
	for _, p := range products {
		_, err = db.GetDB().Exec(
			"INSERT INTO products (name, category, price, stock) VALUES (?, ?, ?, ?)",
			p.name, p.category, p.price, p.stock,
		)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			break
		}
	}
	fmt.Println("  ✓ 准备测试数据完成")

	// 4.1 测试 WHERE 条件
	rows, err = db.GetDB().Query("SELECT name, price FROM products WHERE category = ? AND price > ?", "手机", 5000)
	if err != nil {
		addResult("高级查询", "WHERE条件过滤", false, err.Error())
		fmt.Printf("  ✗ WHERE条件测试失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("高级查询", "WHERE条件过滤", true, fmt.Sprintf("返回 %d 行", count))
		fmt.Printf("  ✓ WHERE条件测试成功(返回%d行)\n", count)
	}

	// 4.2 测试排序
	rows, err = db.GetDB().Query("SELECT name, price FROM products ORDER BY price DESC LIMIT 3")
	if err != nil {
		addResult("高级查询", "ORDER BY排序", false, err.Error())
		fmt.Printf("  ✗ ORDER BY测试失败: %v\n", err)
	} else {
		defer rows.Close()
		names := []string{}
		for rows.Next() {
			var name string
			var price float64
			rows.Scan(&name, &price)
			names = append(names, name)
		}
		addResult("高级查询", "ORDER BY排序", true, fmt.Sprintf("前3: %v", names))
		fmt.Printf("  ✓ ORDER BY测试成功(价格最高3个: %v)\n", names)
	}

	// 4.3 测试聚合函数
	rows, err = db.GetDB().Query("SELECT category, COUNT(*) as count, AVG(price) as avg_price FROM products GROUP BY category")
	if err != nil {
		addResult("高级查询", "GROUP BY聚合", false, err.Error())
		fmt.Printf("  ✗ GROUP BY测试失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		addResult("高级查询", "GROUP BY聚合", true, fmt.Sprintf("%d 个分类", count))
		fmt.Printf("  ✓ GROUP BY聚合测试成功(%d个分类)\n", count)
	}

	// 4.4 测试 LIKE
	rows, err = db.GetDB().Query("SELECT name FROM products WHERE name LIKE ?", "%Pro%")
	if err != nil {
		addResult("高级查询", "LIKE模糊查询", false, err.Error())
		fmt.Printf("  ✗ LIKE测试失败: %v\n", err)
	} else {
		defer rows.Close()
		names := []string{}
		for rows.Next() {
			var name string
			rows.Scan(&name)
			names = append(names, name)
		}
		addResult("高级查询", "LIKE模糊查询", true, fmt.Sprintf("匹配: %v", names))
		fmt.Printf("  ✓ LIKE模糊查询成功(含'Pro'的产品: %v)\n", names)
	}

	// ==================== 5. 空值和特殊字符测试 ====================
	fmt.Println("\n=== 5. 空值和特殊字符测试 ===")

	// 5.1 插入含特殊字符的数据
	_, err = db.GetDB().Exec(
		"INSERT INTO products (name, category, price, stock) VALUES (?, ?, ?, ?)",
		`Special 'quoted' "value" \test`, "测试", 999.99, 10,
	)
	if err != nil {
		addResult("特殊字符", "插入含特殊字符数据", false, err.Error())
		fmt.Printf("  ✗ 特殊字符插入失败: %v\n", err)
	} else {
		addResult("特殊字符", "插入含特殊字符数据", true, "")
		fmt.Println("  ✓ 特殊字符插入成功")
	}

	// 5.2 插入NULL值
	_, err = db.GetDB().Exec(
		"INSERT INTO products (name, category, price, stock) VALUES (?, NULL, ?, ?)",
		"无分类产品", 500.00, 5,
	)
	if err != nil {
		addResult("特殊字符", "插入NULL值", false, err.Error())
		fmt.Printf("  ✗ NULL值插入失败: %v\n", err)
	} else {
		addResult("特殊字符", "插入NULL值", true, "")
		fmt.Println("  ✓ NULL值插入成功")
	}

	// ==================== 6. 错误处理测试 ====================
	fmt.Println("\n=== 6. 错误处理测试 ===")

	// 6.1 查询不存在的表
	_, err = db.GetDB().Query("SELECT * FROM non_existent_table")
	if err != nil {
		addResult("错误处理", "查询不存在的表", true, "正确返回错误")
		fmt.Println("  ✓ 查询不存在的表: 正确返回错误")
	} else {
		addResult("错误处理", "查询不存在的表", false, "应该失败但成功了")
		fmt.Println("  ✗ 查询不存在的表: 应该失败但成功了")
	}

	// 6.2 SQL语法错误
	_, err = db.GetDB().Query("SELECT * FROM products WHERE") // 不完整的SQL
	if err != nil {
		addResult("错误处理", "SQL语法错误", true, "正确返回错误")
		fmt.Println("  ✓ SQL语法错误: 正确返回错误")
	} else {
		addResult("错误处理", "SQL语法错误", false, "应该失败但成功了")
		fmt.Println("  ✗ SQL语法错误: 应该失败但成功了")
	}

	// 6.3 插入重复值(唯一约束)
	_, err = db.GetDB().Exec(`
		CREATE TABLE IF NOT EXISTS unique_test (
			id INT PRIMARY KEY,
			value VARCHAR(50) UNIQUE
		)
	`)
	if err == nil {
		db.GetDB().Exec("INSERT INTO unique_test (id, value) VALUES (1, 'test')")
		_, err = db.GetDB().Exec("INSERT INTO unique_test (id, value) VALUES (2, 'test')")
		if err != nil {
			addResult("错误处理", "唯一约束冲突", true, "正确返回错误")
			fmt.Println("  ✓ 唯一约束冲突: 正确返回错误")
		} else {
			addResult("错误处理", "唯一约束冲突", false, "应该失败但成功了")
			fmt.Println("  ✗ 唯一约束冲突: 应该失败但成功了")
		}
	}

	// ==================== 7. 清理测试数据 ====================
	fmt.Println("\n=== 7. 清理测试数据 ===")
	_, err = db.GetDB().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDB))
	if err != nil {
		fmt.Printf("  ! 清理失败: %v\n", err)
	} else {
		fmt.Println("  ✓ 清理测试数据成功")
	}

	// ==================== 8. 生成测试报告 ====================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                 高级功能测试报告")
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
