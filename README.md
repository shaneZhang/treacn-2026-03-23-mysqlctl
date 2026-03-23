# mysqlctl

一个使用 Go 和 Cobra 框架开发的强大 MySQL 命令行客户端工具。

## 功能列表

本工具提供 **24 个命令**，涵盖数据库操作的各个方面：

### 连接管理

| 命令 | 说明 |
|------|------|
| `connect` | 连接 MySQL 服务器 |
| `disconnect` | 断开 MySQL 连接 |
| `status` | 显示连接状态 |

### 数据库操作

| 命令 | 说明 |
|------|------|
| `databases` | 列出所有数据库 |
| `use` | 切换数据库 |
| `create-database` | 创建数据库 |
| `drop-database` | 删除数据库 |

### 表操作

| 命令 | 说明 |
|------|------|
| `tables` | 列出当前数据库所有表 |
| `describe` | 查看表结构 |
| `show-create` | 查看建表 DDL |

### 数据查询

| 命令 | 说明 |
|------|------|
| `query` | 执行任意 SQL 查询 |
| `select` | 带选项的数据查询（支持分页、排序、条件） |

### 数据操作

| 命令 | 说明 |
|------|------|
| `insert` | 插入数据 |
| `update` | 更新数据 |
| `delete` | 删除数据 |

### 索引管理

| 命令 | 说明 |
|------|------|
| `show-index` | 查看表的索引 |
| `create-index` | 创建索引 |

### 用户权限管理

| 命令 | 说明 |
|------|------|
| `users` | 查看所有用户 |
| `create-user` | 创建用户 |
| `grant` | 授权 |

### 监控

| 命令 | 说明 |
|------|------|
| `processlist` | 显示运行中的进程 |
| `status` | 显示服务器状态 |

### 工具

| 命令 | 说明 |
|------|------|
| `backup` | 备份数据库（需要 mysqldump） |
| `source` | 执行 SQL 文件 |

---

## 安装

### 方式一：二进制安装

```bash
# 克隆项目
git clone https://github.com/your-repo/mysqlctl.git
cd mysqlctl

# 构建
go build -o mysqlctl .

# 添加到 PATH
sudo mv mysqlctl /usr/local/bin/
```

### 方式二：使用 Go install

```bash
go install github.com/your-repo/mysqlctl@latest
```

---

## 快速开始

### 1. 连接数据库

```bash
# 基本连接
mysqlctl connect -h localhost -u root -p

# 指定数据库连接
mysqlctl connect -h localhost -u root -p -d mydb

# 指定端口
mysqlctl connect -h localhost -P 3307 -u root -p
```

### 2. 查看数据库

```bash
# 查看所有数据库
mysqlctl databases

# 切换数据库
mysqlctl use mydb
```

### 3. 表操作

```bash
# 查看所有表
mysqlctl tables

# 查看表结构
mysqlctl describe users

# 查看建表语句
mysqlctl show-create users
```

### 4. 数据查询

```bash
# 执行原始 SQL
mysqlctl query "SELECT * FROM users LIMIT 10"

# 使用 select 命令（推荐）
mysqlctl select --table users --columns "id,name,email"
mysqlctl select --table users --where "status='active'" --limit 10
mysqlctl select --table users --order-by "created_at DESC" --limit 20
```

### 5. 数据操作

```bash
# 插入数据
mysqlctl insert --table users --set "name='张三',email='zhangsan@example.com'"
mysqlctl insert --table users --values "'李四','lisi@example.com'"

# 更新数据
mysqlctl update --table users --set "status='inactive'" --where "id=1"

# 删除数据
mysqlctl delete --table users --where "id=1"
```

### 6. 索引操作

```bash
# 查看索引
mysqlctl show-index users

# 创建索引
mysqlctl create-index --table users --columns "email"
mysqlctl create-index --table users --columns "status,created_at" --name "idx_status_date"
mysqlctl create-index --table users --columns "email" --unique
```

### 7. 用户管理

```bash
# 查看所有用户
mysqlctl users

# 创建用户
mysqlctl create-user newuser --password "password123"
mysqlctl create-user admin --password "securepass" --host "localhost"

# 授权
mysqlctl grant --user newuser --privileges SELECT,INSERT,UPDATE --database mydb
mysqlctl grant --user admin --privileges ALL --database mydb --host localhost
```

### 8. 监控

```bash
# 查看进程列表
mysqlctl processlist

# 查看服务器状态
mysqlctl status
mysqlctl status --extended
```

### 9. 备份还原

```bash
# 备份数据库
mysqlctl backup --database mydb
mysqlctl backup --database mydb --output "backup.sql"

# 执行 SQL 文件
mysqlctl source backup.sql
mysqlctl source --file "backup.sql"
```

---

## 命令详细说明

### connect

连接 MySQL 服务器。

```bash
mysqlctl connect [flags]
```

**参数：**

| 短参数 | 长参数 | 说明 | 默认值 |
|--------|--------|------|--------|
| `-h` | `--host` | MySQL 服务器地址 | localhost |
| `-P` | `--port` | MySQL 端口 | 3306 |
| `-u` | `--user` | 用户名 | root |
| `-p` | `--password` | 密码 | (空) |
| `-d` | `--database` | 默认数据库 | (空) |

---

### query

执行任意 SQL 查询。

```bash
mysqlctl query "SELECT * FROM table WHERE condition"
```

**示例：**

```bash
mysqlctl query "SELECT COUNT(*) FROM users"
mysqlctl query "SHOW TABLES"
```

---

### select

带选项的数据查询。

```bash
mysqlctl select [flags]
```

**参数：**

| 短参数 | 长参数 | 说明 |
|--------|--------|------|
| `-t` | `--table` | 表名（必填） |
| `-c` | `--columns` | 列名，多个用逗号分隔 | * |
| `-w` | `--where` | WHERE 条件 |
| `-o` | `--order-by` | ORDER BY 子句 |
| `-l` | `--limit` | 限制返回行数 |
| `-s` | `--offset` | 偏移量（分页用） |

**示例：**

```bash
# 查询前 10 条
mysqlctl select --table users --limit 10

# 分页查询（第二页，每页 20 条）
mysqlctl select --table users --limit 20 --offset 20

# 条件查询
mysqlctl select --table users --where "status='active'" --order-by "created_at DESC"
```

---

### insert

插入数据。

```bash
mysqlctl insert [flags]
```

**参数：**

| 短参数 | 长参数 | 说明 |
|--------|--------|------|
| `-t` | `--table` | 表名（必填） |
| `-v` | `--values` | 值：'val1', 'val2', ... |
| `-s` | `--set` | 列=值对：col1='val1', col2='val2' |

**示例：**

```bash
# 使用 --set 语法
mysqlctl insert --table users --set "name='张三',email='test@example.com'"

# 使用 --values 语法
mysqlctl insert --table users --values "'张三','test@example.com'"
```

---

### update

更新数据。

```bash
mysqlctl update [flags]
```

**参数：**

| 短参数 | 长参数 | 说明 |
|--------|--------|------|
| `-t` | `--table` | 表名（必填） |
| `-s` | `--set` | 列=值对（必填） |
| `-w` | `--where` | WHERE 条件 |

**示例：**

```bash
mysqlctl update --table users --set "name='李四'" --where "id=1"
mysqlctl update --table users --set "status='inactive'" --where "created_at < '2024-01-01'"
```

---

### delete

删除数据。

```bash
mysqlctl delete [flags]
```

**参数：**

| 短参数 | 长参数 | 说明 |
|--------|--------|------|
| `-t` | `--table` | 表名（必填） |
| `-w` | `--where` | WHERE 条件（必填） |

**示例：**

```bash
mysqlctl delete --table users --where "id=1"
mysqlctl delete --table users --where "status='inactive' AND created_at < '2024-01-01'"
```

---

### create-index

创建索引。

```bash
mysqlctl create-index [flags]
```

**参数：**

| 短参数 | 长参数 | 说明 |
|--------|--------|------|
| `-t` | `--table` | 表名（必填） |
| `-n` | `--name` | 索引名（可选） |
| `-c` | `--columns` | 列名，多个用逗号分隔（必填） |
| `-u` | `--unique` | 是否唯一索引 |

**示例：**

```bash
mysqlctl create-index --table users --columns "email"
mysqlctl create-index --table orders --columns "user_id,status" --name "idx_user_status"
mysqlctl create-index --table users --columns "email" --unique
```

---

### grant

授予用户权限。

```bash
mysqlctl grant [flags]
```

**参数：**

| 短参数 | 长参数 | 说明 |
|--------|--------|------|
| `-u` | `--user` | 用户名（必填） |
| `-h` | `--host` | 主机 | % |
| `-p` | `--privileges` | 权限：SELECT, INSERT, UPDATE, DELETE, ALL 等（必填） |
| `-d` | `--database` | 数据库名 | * |
| `-t` | `--table` | 表名（可选） |

**示例：**

```bash
mysqlctl grant --user appuser --privileges SELECT,INSERT,UPDATE,DELETE --database mydb
mysqlctl grant --user admin --privileges ALL --database mydb --host localhost
```

---

### backup

备份数据库（需要系统安装 mysqldump）。

```bash
mysqlctl backup [flags]
```

**参数：**

| 短参数 | 长参数 | 说明 |
|--------|--------|------|
| `-d` | `--database` | 数据库名（必填） |
| `-o` | `--output` | 输出文件路径 |

**示例：**

```bash
mysqlctl backup --database mydb
mysqlctl backup --database mydb --output "/tmp/backup_$(date +%Y%m%d).sql"
```

---

## 注意事项

1. **连接安全**：密码作为命令行参数传递时需注意安全，生产环境建议使用环境变量或交互式输入
2. **SQL 注入**：使用 `--where`、`--set` 等参数时，确保输入已正确转义，防止 SQL 注入
3. **备份功能**：需要系统已安装 `mysqldump` 工具
4. **权限**：部分操作（如创建用户、授权）需要 MySQL 管理员权限

---

## 许可证

MIT License
