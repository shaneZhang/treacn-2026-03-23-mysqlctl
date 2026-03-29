package main

import (
	"fmt"
	"mysqlctl/internal/db"
)

func main() {
	fmt.Println("=== MySQLCTL 综合测试 ===")

	// 1. 测试：成功连接
	fmt.Println("\n1. 测试正确连接...")
	err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if err != nil {
		fmt.Printf("  ✗ 失败: %v\n", err)
	} else {
		fmt.Println("  ✓ 成功")
		db.Disconnect()
	}

	// 2. 测试：错误密码
	fmt.Println("\n2. 测试错误密码连接...")
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "wrongpass", "")
	if err != nil {
		fmt.Printf("  ✓ 正确拒绝访问\n")
	} else {
		fmt.Println("  ✗ 应该失败但成功了")
		db.Disconnect()
	}

	// 3. 测试：连接并指定数据库
	fmt.Println("\n3. 测试连接并指定初始数据库(mysql)...")
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "mysql")
	if err != nil {
		fmt.Printf("  ✗ 失败: %v\n", err)
	} else {
		fmt.Println("  ✓ 成功连接并指定数据库")
		db.Disconnect()
	}

	// 4. 测试：获取数据库列表
	fmt.Println("\n4. 测试获取数据库列表...")
	err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
	if err != nil {
		fmt.Printf("  ✗ 连接失败: %v\n", err)
		return
	}
	defer db.Disconnect()

	rows, err := db.GetDB().Query("SHOW DATABASES")
	if err != nil {
		fmt.Printf("  ✗ 查询失败: %v\n", err)
	} else {
		defer rows.Close()
		count := 0
		var hasInfoSchema, hasMySQL bool
		for rows.Next() {
			var name string
			rows.Scan(&name)
			count++
			if name == "information_schema" {
				hasInfoSchema = true
			}
			if name == "mysql" {
				hasMySQL = true
			}
		}
		fmt.Printf("  ✓ 获取到 %d 个数据库", count)
		if hasInfoSchema && hasMySQL {
			fmt.Println(" (包含系统数据库)")
		} else {
			fmt.Println()
		}
	}

	fmt.Println("\n=== 测试完成 ===")
}
