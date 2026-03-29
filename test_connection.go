package main

import (
"fmt"
"mysqlctl/internal/db"
)

func main() {
// 测试连接
fmt.Println("=== 测试连接功能 ===")

// 1. 测试正确连接
err := db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
if err != nil {
fmt.Printf("连接失败: %v\n", err)
} else {
fmt.Println("✓ 使用正确凭据连接成功")
}

// 测试状态
connected, dbName := db.GetStatus()
if connected {
fmt.Println("✓ 状态检查: 已连接")
}

// 断开连接
db.Disconnect()
connected, _ = db.GetStatus()
if !connected {
fmt.Println("✓ 断开连接成功")
}

// 2. 测试数据库列表
fmt.Println("\n=== 测试 databases 命令 ===")
err = db.Connect("192.168.31.210", 3306, "zhangyuqing", "zhangyuqing", "")
if err != nil {
fmt.Printf("连接失败: %v\n", err)
return
}
defer db.Disconnect()

rows, err := db.GetDB().Query("SHOW DATABASES")
ipackage main

import (
"fmt"
"mysqlctl/internal/db"
)

func main() {
// 测试连接
fmt.Println("=== 测试连n(
import (
???"fmt"
fo"myss.)

func main() {
// ?ing
// 测试?&nfmt.fmt.Printf
// 1. 测试正确连接
err := db.Co???err := db.Connect("192.
}if err != nil {
fnection.go 2>&1