# 数据库迁移 (Database Migrations)

## 概述

数据库迁移是管理数据库 schema 变更的版本控制系统，确保数据库结构在不同环境间保持一致。

## 为什么需要迁移？

```
❌ 手动管理 Schema:
- 开发者 A 添加了表，开发者 B 不知道
- 生产环境 schema 和测试环境不一致
- 无法追踪历史变更
- 回滚困难

✅ 使用迁移:
- 每个变更都有版本记录
- 可以在任何环境重建数据库
- 支持回滚
- 团队协作友好
```

## golang-migrate

### 安装

```bash
# CLI 工具
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 或使用 Homebrew
brew install golang-migrate

# Go 库
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
```

### 目录结构

```
project/
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_add_email_to_users.up.sql
│   ├── 000002_add_email_to_users.down.sql
│   ├── 000003_create_posts_table.up.sql
│   └── 000003_create_posts_table.down.sql
└── main.go
```

### 创建迁移文件

```bash
# 创建新迁移
migrate create -ext sql -dir migrations -seq create_users_table

# 生成文件:
# migrations/000001_create_users_table.up.sql
# migrations/000001_create_users_table.down.sql
```

### 迁移文件示例

```sql
-- migrations/000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- migrations/000001_create_users_table.down.sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

```sql
-- migrations/000002_add_status_to_users.up.sql
ALTER TABLE users
ADD COLUMN status VARCHAR(20) DEFAULT 'active';

CREATE INDEX idx_users_status ON users(status);

-- migrations/000002_add_status_to_users.down.sql
DROP INDEX IF EXISTS idx_users_status;
ALTER TABLE users DROP COLUMN status;
```

```sql
-- migrations/000003_create_posts_table.up.sql
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'draft',
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_status ON posts(status);

-- migrations/000003_create_posts_table.down.sql
DROP INDEX IF EXISTS idx_posts_status;
DROP INDEX IF EXISTS idx_posts_author;
DROP TABLE IF EXISTS posts;
```

### CLI 使用

```bash
# 数据库 URL
export DATABASE_URL="postgres://user:password@localhost:5432/testdb?sslmode=disable"

# 执行所有迁移
migrate -path migrations -database "$DATABASE_URL" up

# 执行指定数量的迁移
migrate -path migrations -database "$DATABASE_URL" up 2

# 回滚一个迁移
migrate -path migrations -database "$DATABASE_URL" down 1

# 回滚所有迁移
migrate -path migrations -database "$DATABASE_URL" down

# 查看当前版本
migrate -path migrations -database "$DATABASE_URL" version

# 强制设置版本（不执行迁移，修复脏状态）
migrate -path migrations -database "$DATABASE_URL" force 3

# 删除所有（包括 schema_migrations 表）
migrate -path migrations -database "$DATABASE_URL" drop
```

### Go 代码中使用

```go
package main

import (
    "database/sql"
    "log"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    _ "github.com/lib/pq"
)

func runMigrations(db *sql.DB) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return err
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",  // 迁移文件目录
        "postgres",           // 数据库名
        driver,
    )
    if err != nil {
        return err
    }

    // 执行所有向上迁移
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    log.Println("Migrations completed")
    return nil
}

// 带版本控制
func migrateToVersion(db *sql.DB, version uint) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return err
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",
        "postgres",
        driver,
    )
    if err != nil {
        return err
    }

    return m.Migrate(version)
}

// 回滚
func rollback(db *sql.DB, steps int) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return err
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",
        "postgres",
        driver,
    )
    if err != nil {
        return err
    }

    return m.Steps(-steps)
}
```

### 嵌入迁移文件

```go
package main

import (
    "embed"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func runEmbeddedMigrations(db *sql.DB) error {
    // 从嵌入的文件系统读取迁移
    source, err := iofs.New(migrationsFS, "migrations")
    if err != nil {
        return err
    }

    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return err
    }

    m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
    if err != nil {
        return err
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    return nil
}
```

## goose

另一个流行的迁移工具。

### 安装

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### 创建迁移

```bash
goose -dir migrations create create_users_table sql

# 生成:
# migrations/20240101120000_create_users_table.sql
```

### SQL 迁移文件

```sql
-- migrations/20240101120000_create_users_table.sql

-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS users;
```

### Go 迁移文件

```go
// migrations/20240101130000_seed_data.go

package migrations

import (
    "context"
    "database/sql"

    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigrationContext(upSeedData, downSeedData)
}

func upSeedData(ctx context.Context, tx *sql.Tx) error {
    _, err := tx.ExecContext(ctx, `
        INSERT INTO users (name, email) VALUES
        ('Admin', 'admin@example.com'),
        ('User', 'user@example.com')
    `)
    return err
}

func downSeedData(ctx context.Context, tx *sql.Tx) error {
    _, err := tx.ExecContext(ctx, `
        DELETE FROM users WHERE email IN ('admin@example.com', 'user@example.com')
    `)
    return err
}
```

### CLI 使用

```bash
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="user=user password=password dbname=testdb sslmode=disable"

# 执行迁移
goose -dir migrations up

# 回滚
goose -dir migrations down

# 查看状态
goose -dir migrations status

# 指定版本
goose -dir migrations up-to 20240101120000
```

### Go 代码中使用

```go
package main

import (
    "database/sql"
    "log"

    "github.com/pressly/goose/v3"
    _ "github.com/lib/pq"
)

func runGooseMigrations(db *sql.DB) error {
    goose.SetDialect("postgres")

    if err := goose.Up(db, "migrations"); err != nil {
        return err
    }

    log.Println("Goose migrations completed")
    return nil
}
```

## 迁移最佳实践

### 1. 迁移必须可逆

```sql
-- ✅ 好的迁移 - 可以完全回滚

-- up.sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- down.sql
ALTER TABLE users DROP COLUMN phone;
```

### 2. 避免破坏性变更

```sql
-- ❌ 危险 - 直接删除列可能丢失数据
ALTER TABLE users DROP COLUMN phone;

-- ✅ 安全的方法:
-- 1. 先停止写入该列
-- 2. 备份数据
-- 3. 在低峰期执行删除
-- 4. 保留回滚脚本
```

### 3. 分步重命名列

```sql
-- ❌ 直接重命名可能导致应用错误
ALTER TABLE users RENAME COLUMN name TO full_name;

-- ✅ 分步执行:

-- 步骤 1: 添加新列
ALTER TABLE users ADD COLUMN full_name VARCHAR(100);

-- 步骤 2: 复制数据
UPDATE users SET full_name = name;

-- 步骤 3: 更新应用代码使用 full_name

-- 步骤 4: 删除旧列（确认无误后）
ALTER TABLE users DROP COLUMN name;
```

### 4. 大表迁移

```sql
-- ❌ 大表添加带默认值的列会锁表
ALTER TABLE large_table ADD COLUMN status VARCHAR(20) DEFAULT 'active';

-- ✅ PostgreSQL 11+ 安全方式（不锁表）
ALTER TABLE large_table ADD COLUMN status VARCHAR(20);
ALTER TABLE large_table ALTER COLUMN status SET DEFAULT 'active';
UPDATE large_table SET status = 'active' WHERE status IS NULL;

-- ✅ 分批更新
DO $$
DECLARE
    batch_size INT := 10000;
    affected INT;
BEGIN
    LOOP
        UPDATE large_table
        SET status = 'active'
        WHERE id IN (
            SELECT id FROM large_table
            WHERE status IS NULL
            LIMIT batch_size
        );

        GET DIAGNOSTICS affected = ROW_COUNT;
        EXIT WHEN affected = 0;

        COMMIT;
        PERFORM pg_sleep(0.1);  -- 短暂休息
    END LOOP;
END $$;
```

### 5. 添加索引

```sql
-- ❌ 普通创建索引会锁表
CREATE INDEX idx_users_email ON users(email);

-- ✅ 使用 CONCURRENTLY（不锁表，但更慢）
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- 注意: CONCURRENTLY 不能在事务中使用
```

### 6. 外键约束

```sql
-- ❌ 添加外键会锁定两个表
ALTER TABLE posts ADD CONSTRAINT fk_author
    FOREIGN KEY (author_id) REFERENCES users(id);

-- ✅ 分步添加
-- 步骤 1: 添加无效约束（不验证）
ALTER TABLE posts ADD CONSTRAINT fk_author
    FOREIGN KEY (author_id) REFERENCES users(id)
    NOT VALID;

-- 步骤 2: 异步验证
ALTER TABLE posts VALIDATE CONSTRAINT fk_author;
```

## 迁移测试

```go
func TestMigrations(t *testing.T) {
    // 创建测试数据库
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    // 执行所有向上迁移
    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",
        "postgres",
        driver,
    )
    require.NoError(t, err)

    err = m.Up()
    require.NoError(t, err)

    // 验证表存在
    var exists bool
    err = db.QueryRow(`
        SELECT EXISTS (
            SELECT FROM information_schema.tables
            WHERE table_name = 'users'
        )
    `).Scan(&exists)
    require.NoError(t, err)
    assert.True(t, exists)

    // 测试回滚
    err = m.Down()
    require.NoError(t, err)
}
```

## CI/CD 集成

### Makefile

```makefile
.PHONY: migrate-up migrate-down migrate-create

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-force:
	@read -p "Enter version: " version; \
	migrate -path migrations -database "$(DATABASE_URL)" force $$version
```

### GitHub Actions

```yaml
# .github/workflows/migrate.yml
name: Database Migration

on:
  push:
    branches: [main]
    paths:
      - 'migrations/**'

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install migrate
        run: |
          curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
          sudo mv migrate /usr/local/bin/

      - name: Run migrations
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          migrate -path migrations -database "$DATABASE_URL" up
```

## 常见问题

### 脏状态修复

```bash
# 当迁移中断时，可能留下脏状态
# 查看当前状态
migrate -path migrations -database "$DATABASE_URL" version

# 强制设置版本（跳过当前版本的执行）
migrate -path migrations -database "$DATABASE_URL" force VERSION
```

### 迁移顺序冲突

```
当多人同时创建迁移时：
- 开发者 A 创建 000004_xxx.sql
- 开发者 B 也创建 000004_yyy.sql

解决方案：
- 使用时间戳命名: 20240101120000_xxx.sql
- 合并前重新编号
- 使用 goose（支持时间戳）
```

**下一节**：[练习](./exercises.md) - 数据库操作练习
