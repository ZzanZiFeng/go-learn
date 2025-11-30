# PostgreSQL 安装与配置

## 概述

PostgreSQL 是功能强大的开源关系数据库，是 Go 后端开发的首选数据库之一。

## 安装方式

### macOS

```bash
# Homebrew 安装
brew install postgresql@15

# 启动服务
brew services start postgresql@15

# 停止服务
brew services stop postgresql@15

# 查看状态
brew services info postgresql@15

# 创建默认数据库
createdb

# 进入 psql
psql
```

### Linux (Ubuntu/Debian)

```bash
# 添加 PostgreSQL 官方仓库
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -

# 安装
sudo apt-get update
sudo apt-get install postgresql-15

# 启动服务
sudo systemctl start postgresql
sudo systemctl enable postgresql

# 切换到 postgres 用户
sudo -u postgres psql
```

### Windows

1. 下载安装程序：https://www.postgresql.org/download/windows/
2. 运行安装向导
3. 设置密码和端口（默认 5432）
4. 使用 pgAdmin 或 psql 连接

### Docker（推荐开发环境）

```bash
# 运行 PostgreSQL 容器
docker run --name postgres \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  -v postgres_data:/var/lib/postgresql/data \
  -d postgres:15

# 查看日志
docker logs postgres

# 进入 psql
docker exec -it postgres psql -U user -d testdb

# 停止容器
docker stop postgres

# 启动容器
docker start postgres

# 删除容器
docker rm -f postgres
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    container_name: postgres
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: testdb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d testdb"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

```bash
# 启动
docker-compose up -d

# 停止
docker-compose down

# 查看日志
docker-compose logs -f postgres
```

## 初始配置

### 创建用户和数据库

```sql
-- 进入 psql
psql -U postgres

-- 创建用户
CREATE USER myapp WITH PASSWORD 'secret123';

-- 创建数据库
CREATE DATABASE myapp_dev OWNER myapp;
CREATE DATABASE myapp_test OWNER myapp;

-- 授权
GRANT ALL PRIVILEGES ON DATABASE myapp_dev TO myapp;
GRANT ALL PRIVILEGES ON DATABASE myapp_test TO myapp;

-- 查看数据库列表
\l

-- 查看用户列表
\du

-- 连接到数据库
\c myapp_dev

-- 查看表列表
\dt

-- 退出
\q
```

### 配置文件

```bash
# 查找配置文件位置
psql -U postgres -c "SHOW config_file;"
# 通常在 /etc/postgresql/15/main/postgresql.conf

# 查找 HBA 配置
psql -U postgres -c "SHOW hba_file;"
# 通常在 /etc/postgresql/15/main/pg_hba.conf
```

### postgresql.conf 常用配置

```conf
# 连接设置
listen_addresses = '*'          # 监听所有地址（开发环境）
port = 5432                     # 端口
max_connections = 100           # 最大连接数

# 内存设置
shared_buffers = 256MB          # 共享缓冲区（物理内存的 25%）
effective_cache_size = 768MB    # 有效缓存大小（物理内存的 75%）
work_mem = 4MB                  # 排序/哈希操作内存
maintenance_work_mem = 64MB     # 维护操作内存

# 日志设置
log_destination = 'stderr'
logging_collector = on
log_directory = 'log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_statement = 'all'           # 记录所有 SQL（开发环境）

# 性能设置
random_page_cost = 1.1          # SSD 优化
effective_io_concurrency = 200  # SSD 优化
```

### pg_hba.conf 认证配置

```conf
# TYPE  DATABASE        USER            ADDRESS                 METHOD

# 本地连接
local   all             all                                     trust
host    all             all             127.0.0.1/32            md5
host    all             all             ::1/128                 md5

# Docker 网络（开发环境）
host    all             all             172.0.0.0/8             md5

# 远程连接（生产环境需要更严格的配置）
# host    all             all             0.0.0.0/0               md5
```

## 连接字符串

### 标准格式

```
postgresql://[user[:password]@][host][:port][/database][?param=value&...]
```

### 示例

```bash
# 基本连接
postgresql://user:password@localhost:5432/testdb

# 带 SSL
postgresql://user:password@localhost:5432/testdb?sslmode=require

# 完整参数
postgresql://user:password@localhost:5432/testdb?sslmode=disable&connect_timeout=10&application_name=myapp
```

### Go 中使用

```go
package main

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/lib/pq"
)

func main() {
    // 方式 1: 连接字符串
    connStr := "postgres://user:password@localhost:5432/testdb?sslmode=disable"

    // 方式 2: 环境变量
    connStr = os.Getenv("DATABASE_URL")

    // 方式 3: DSN 格式
    connStr = "host=localhost port=5432 user=user password=password dbname=testdb sslmode=disable"

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        panic(err)
    }

    fmt.Println("Connected!")
}
```

## SSL/TLS 配置

### SSL 模式

| 模式 | 说明 |
|------|------|
| disable | 禁用 SSL |
| allow | 先尝试非 SSL，失败再尝试 SSL |
| prefer | 先尝试 SSL，失败再尝试非 SSL |
| require | 必须 SSL，但不验证证书 |
| verify-ca | SSL + 验证服务器证书是否由可信 CA 签发 |
| verify-full | SSL + 验证证书 + 验证主机名 |

### 生产环境配置

```go
// 生产环境应使用 verify-full
connStr := "postgres://user:password@prod-db.example.com:5432/proddb?sslmode=verify-full"
```

## GUI 工具

### pgAdmin 4

官方 GUI 工具：https://www.pgadmin.org/

```bash
# macOS
brew install --cask pgadmin4

# Docker
docker run -p 5050:80 \
  -e PGADMIN_DEFAULT_EMAIL=admin@admin.com \
  -e PGADMIN_DEFAULT_PASSWORD=admin \
  -d dpage/pgadmin4
```

### DBeaver

通用数据库 GUI：https://dbeaver.io/

```bash
# macOS
brew install --cask dbeaver-community
```

### TablePlus

现代 GUI（macOS/Windows）：https://tableplus.com/

### DataGrip

JetBrains IDE（付费）：https://www.jetbrains.com/datagrip/

## 常用 psql 命令

```bash
# 连接
psql -h localhost -p 5432 -U user -d testdb

# 执行 SQL 文件
psql -U user -d testdb -f script.sql

# 执行单条命令
psql -U user -d testdb -c "SELECT * FROM users"

# 导出数据
pg_dump -U user testdb > backup.sql
pg_dump -U user -F c testdb > backup.dump  # 自定义格式

# 导入数据
psql -U user testdb < backup.sql
pg_restore -U user -d testdb backup.dump
```

### psql 内部命令

```sql
\l          -- 列出数据库
\c dbname   -- 连接数据库
\dt         -- 列出表
\dt+        -- 详细表信息
\d table    -- 描述表结构
\di         -- 列出索引
\du         -- 列出用户
\df         -- 列出函数
\dv         -- 列出视图
\timing     -- 显示执行时间
\x          -- 扩展显示模式
\q          -- 退出
```

## 项目 Docker 配置

使用项目的 `infra/docker-compose.yml`：

```bash
cd /path/to/go-learn/infra

# 启动 PostgreSQL
docker-compose up -d postgres

# 查看状态
docker-compose ps

# 连接
docker-compose exec postgres psql -U user -d testdb

# 查看日志
docker-compose logs -f postgres
```

## 验证安装

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"
)

func main() {
    connStr := "postgres://user:password@localhost:5432/testdb?sslmode=disable"

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Failed to open:", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal("Failed to ping:", err)
    }

    var version string
    err = db.QueryRow("SELECT version()").Scan(&version)
    if err != nil {
        log.Fatal("Failed to query:", err)
    }

    fmt.Println("PostgreSQL version:")
    fmt.Println(version)
}
```

运行：

```bash
go mod init dbtest
go get github.com/lib/pq
go run main.go
```

输出：

```
PostgreSQL version:
PostgreSQL 15.4 on aarch64-apple-darwin22.1.0, compiled by ...
```

**下一节**：[database/sql](./03-database-sql.md) - 学习 Go 标准库数据库操作
