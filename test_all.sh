#!/bin/bash

MYSQLCTL="./mysqlctl"
export MYSQL_HOST="192.168.31.112"
export MYSQL_PORT="3306"
export MYSQL_USER="zhangyuqing"
export MYSQL_PASSWORD="zhangyuqing"
TEST_DB="test_mysqlctl_db"
TEST_TABLE="test_table"

echo "=========================================="
echo "MySQLctl 全功能测试"
echo "=========================================="

# 先确保数据库存在（不指定数据库）
echo ""
echo "【准备】确保测试数据库存在"
unset MYSQL_DATABASE
$MYSQLCTL create-database $TEST_DB 2>/dev/null || echo "数据库已存在"

# 设置数据库环境变量
export MYSQL_DATABASE=$TEST_DB

# 1. 测试连接
echo ""
echo "【测试1】connect - 连接数据库"
$MYSQLCTL connect -H $MYSQL_HOST -P $MYSQL_PORT -u $MYSQL_USER -p $MYSQL_PASSWORD -d $TEST_DB
echo "✓ 连接命令执行成功"

# 2. 测试显示数据库
echo ""
echo "【测试2】databases - 显示所有数据库"
$MYSQLCTL databases

# 3. 测试使用数据库
echo ""
echo "【测试3】use - 切换到测试数据库"
$MYSQLCTL use $TEST_DB

# 4. 测试状态
echo ""
echo "【测试4】status - 显示连接状态"
$MYSQLCTL status

# 5. 测试显示表
echo ""
echo "【测试5】tables - 显示表"
$MYSQLCTL tables

# 6. 测试query创建表
echo ""
echo "【测试6】query - 执行CREATE TABLE语句"
$MYSQLCTL query "CREATE TABLE IF NOT EXISTS $TEST_TABLE (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(100), age INT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)"

# 7. 测试显示表
echo ""
echo "【测试7】tables - 显示表"
$MYSQLCTL tables

# 8. 测试描述表
echo ""
echo "【测试8】describe - 描述表结构"
$MYSQLCTL describe $TEST_TABLE

# 9. 测试插入数据
echo ""
echo "【测试9】insert - 插入数据"
$MYSQLCTL insert --table $TEST_TABLE --set "name='Alice', age=25"
$MYSQLCTL insert --table $TEST_TABLE --set "name='Bob', age=30"
$MYSQLCTL insert --table $TEST_TABLE --set "name='Charlie', age=35"

# 10. 测试查询数据
echo ""
echo "【测试10】select - 查询所有数据"
$MYSQLCTL select --table $TEST_TABLE

# 11. 测试带条件查询
echo ""
echo "【测试11】select - 带WHERE条件查询"
$MYSQLCTL select --table $TEST_TABLE --where "age > 26"

# 12. 测试排序查询
echo ""
echo "【测试12】select - 带ORDER BY查询"
$MYSQLCTL select --table $TEST_TABLE --order-by "age DESC"

# 13. 测试分页查询
echo ""
echo "【测试13】select - 带LIMIT分页查询"
$MYSQLCTL select --table $TEST_TABLE --limit 2

# 14. 测试更新数据
echo ""
echo "【测试14】update - 更新数据"
$MYSQLCTL update --table $TEST_TABLE --set "age=26" --where "name='Alice'"

# 15. 验证更新
echo ""
echo "【测试15】select - 验证更新结果"
$MYSQLCTL select --table $TEST_TABLE

# 16. 测试显示创建表
echo ""
echo "【测试16】show-create - 显示建表语句"
$MYSQLCTL show-create $TEST_TABLE

# 17. 测试创建索引
echo ""
echo "【测试17】create-index - 创建索引"
$MYSQLCTL create-index --table $TEST_TABLE --columns name --name idx_name

# 18. 测试显示索引
echo ""
echo "【测试18】show-index - 显示索引"
$MYSQLCTL show-index $TEST_TABLE

# 19. 测试删除数据
echo ""
echo "【测试19】delete - 删除数据"
$MYSQLCTL delete --table $TEST_TABLE --where "name='Bob'"

# 20. 验证删除
echo ""
echo "【测试20】select - 验证删除结果"
$MYSQLCTL select --table $TEST_TABLE

# 21. 测试显示用户
echo ""
echo "【测试21】users - 显示所有用户"
$MYSQLCTL users

# 22. 测试进程列表
echo ""
echo "【测试22】processlist - 显示进程列表"
$MYSQLCTL processlist

# 23. 测试服务器状态
echo ""
echo "【测试23】server-status - 显示服务器状态"
$MYSQLCTL server-status

# 24. 测试扩展服务器状态
echo ""
echo "【测试24】server-status --extended - 显示扩展服务器状态"
$MYSQLCTL server-status --extended | head -15

# 25. 测试source命令
echo ""
echo "【测试25】source - 执行SQL文件"
SQL_FILE="/tmp/test_source.sql"
echo "INSERT INTO $TEST_TABLE (name, age) VALUES ('David', 40);" > $SQL_FILE
echo "INSERT INTO $TEST_TABLE (name, age) VALUES ('Eve', 45);" >> $SQL_FILE
$MYSQLCTL source $SQL_FILE
$MYSQLCTL select --table $TEST_TABLE

# 26. 清理 - 删除表
echo ""
echo "【测试26】query - 删除表"
$MYSQLCTL query "DROP TABLE IF EXISTS $TEST_TABLE"

# 27. 测试删除数据库
echo ""
echo "【测试27】drop-database - 删除测试数据库"
unset MYSQL_DATABASE
$MYSQLCTL drop-database $TEST_DB

# 28. 测试断开连接
echo ""
echo "【测试28】disconnect - 断开连接"
export MYSQL_DATABASE=$TEST_DB
$MYSQLCTL disconnect

# 29. 验证断开
echo ""
echo "【测试29】status - 验证已断开连接"
unset MYSQL_DATABASE
$MYSQLCTL status

echo ""
echo "=========================================="
echo "所有测试完成！"
echo "=========================================="
