# SQL 基础复习 (SQL Review)

## 概述

在学习 Go 数据库操作之前，先复习 SQL 基础语法。本节是快速参考，不是 SQL 教程。

## 基本 CRUD 操作

### CREATE - 插入数据

```sql
-- 插入单行
INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com');

-- 插入多行
INSERT INTO users (name, email) VALUES
    ('Alice', 'alice@example.com'),
    ('Bob', 'bob@example.com');

-- 插入并返回生成的 ID (PostgreSQL)
INSERT INTO users (name, email)
VALUES ('Alice', 'alice@example.com')
RETURNING id;

-- 插入并返回生成的 ID (MySQL)
INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com');
SELECT LAST_INSERT_ID();
```

### READ - 查询数据

```sql
-- 查询所有列
SELECT * FROM users;

-- 查询指定列
SELECT name, email FROM users;

-- 条件查询
SELECT * FROM users WHERE id = 1;
SELECT * FROM users WHERE name LIKE 'A%';
SELECT * FROM users WHERE created_at > '2024-01-01';

-- 排序
SELECT * FROM users ORDER BY created_at DESC;

-- 分页
SELECT * FROM users ORDER BY id LIMIT 10 OFFSET 20;

-- 聚合
SELECT COUNT(*) FROM users;
SELECT role, COUNT(*) FROM users GROUP BY role;
SELECT role, COUNT(*) FROM users GROUP BY role HAVING COUNT(*) > 5;
```

### UPDATE - 更新数据

```sql
-- 更新单个字段
UPDATE users SET name = 'Alice Smith' WHERE id = 1;

-- 更新多个字段
UPDATE users SET name = 'Alice Smith', email = 'alice.smith@example.com' WHERE id = 1;

-- 条件更新
UPDATE users SET status = 'inactive' WHERE last_login < '2023-01-01';

-- 更新并返回 (PostgreSQL)
UPDATE users SET name = 'New Name' WHERE id = 1 RETURNING *;
```

### DELETE - 删除数据

```sql
-- 删除单行
DELETE FROM users WHERE id = 1;

-- 条件删除
DELETE FROM users WHERE status = 'inactive';

-- 删除所有数据（谨慎！）
DELETE FROM users;

-- 软删除（推荐）
UPDATE users SET deleted_at = NOW() WHERE id = 1;
```

## 表结构管理

### 创建表

```sql
-- 基础表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 完整示例
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    author_id INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'draft',
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- 外键约束
    CONSTRAINT fk_author
        FOREIGN KEY (author_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    -- 检查约束
    CONSTRAINT chk_status
        CHECK (status IN ('draft', 'published', 'archived'))
);

-- 创建索引
CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_created ON posts(created_at DESC);
```

### 修改表

```sql
-- 添加列
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- 修改列
ALTER TABLE users ALTER COLUMN phone TYPE VARCHAR(30);

-- 删除列
ALTER TABLE users DROP COLUMN phone;

-- 添加约束
ALTER TABLE users ADD CONSTRAINT unique_phone UNIQUE (phone);

-- 删除约束
ALTER TABLE users DROP CONSTRAINT unique_phone;

-- 添加索引
CREATE INDEX idx_users_name ON users(name);

-- 删除索引
DROP INDEX idx_users_name;
```

### 删除表

```sql
-- 删除表（如果存在）
DROP TABLE IF EXISTS users;

-- 级联删除（删除依赖）
DROP TABLE IF EXISTS users CASCADE;
```

## JOIN 操作

### 示例数据结构

```sql
-- users 表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100)
);

-- posts 表
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255),
    author_id INTEGER REFERENCES users(id)
);

-- comments 表
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    content TEXT,
    post_id INTEGER REFERENCES posts(id),
    user_id INTEGER REFERENCES users(id)
);
```

### INNER JOIN

```sql
-- 返回两表都有匹配的行
SELECT
    p.id,
    p.title,
    u.name AS author_name
FROM posts p
INNER JOIN users u ON p.author_id = u.id;
```

### LEFT JOIN

```sql
-- 返回左表所有行，右表可能为 NULL
SELECT
    u.id,
    u.name,
    COUNT(p.id) AS post_count
FROM users u
LEFT JOIN posts p ON u.id = p.author_id
GROUP BY u.id, u.name;
```

### RIGHT JOIN

```sql
-- 返回右表所有行，左表可能为 NULL
SELECT
    p.title,
    u.name
FROM posts p
RIGHT JOIN users u ON p.author_id = u.id;
```

### 多表 JOIN

```sql
-- 帖子、作者、评论数
SELECT
    p.id,
    p.title,
    u.name AS author,
    COUNT(c.id) AS comment_count
FROM posts p
INNER JOIN users u ON p.author_id = u.id
LEFT JOIN comments c ON p.id = c.post_id
GROUP BY p.id, p.title, u.name;
```

### 自连接

```sql
-- 员工和经理关系
SELECT
    e.name AS employee,
    m.name AS manager
FROM employees e
LEFT JOIN employees m ON e.manager_id = m.id;
```

## 子查询

### 标量子查询

```sql
-- 返回单个值
SELECT
    name,
    (SELECT COUNT(*) FROM posts WHERE author_id = users.id) AS post_count
FROM users;
```

### IN 子查询

```sql
-- 有帖子的用户
SELECT * FROM users
WHERE id IN (SELECT DISTINCT author_id FROM posts);
```

### EXISTS 子查询

```sql
-- 有帖子的用户（性能更好）
SELECT * FROM users u
WHERE EXISTS (
    SELECT 1 FROM posts p WHERE p.author_id = u.id
);
```

### FROM 子查询

```sql
-- 统计每个作者的平均帖子长度
SELECT
    author_id,
    AVG(post_length) as avg_length
FROM (
    SELECT
        author_id,
        LENGTH(content) as post_length
    FROM posts
) AS post_stats
GROUP BY author_id;
```

## PostgreSQL 特有功能

### RETURNING 子句

```sql
-- INSERT 返回
INSERT INTO users (name, email)
VALUES ('Alice', 'alice@example.com')
RETURNING id, created_at;

-- UPDATE 返回
UPDATE users SET name = 'New Name'
WHERE id = 1
RETURNING *;

-- DELETE 返回
DELETE FROM users WHERE id = 1 RETURNING *;
```

### UPSERT (INSERT ... ON CONFLICT)

```sql
-- 插入或更新
INSERT INTO users (email, name)
VALUES ('alice@example.com', 'Alice')
ON CONFLICT (email)
DO UPDATE SET name = EXCLUDED.name, updated_at = NOW();

-- 插入或忽略
INSERT INTO users (email, name)
VALUES ('alice@example.com', 'Alice')
ON CONFLICT (email) DO NOTHING;
```

### CTE (Common Table Expressions)

```sql
-- WITH 子句
WITH active_users AS (
    SELECT * FROM users WHERE status = 'active'
),
user_posts AS (
    SELECT
        u.id,
        u.name,
        COUNT(p.id) as post_count
    FROM active_users u
    LEFT JOIN posts p ON u.id = p.author_id
    GROUP BY u.id, u.name
)
SELECT * FROM user_posts WHERE post_count > 5;
```

### JSON 操作

```sql
-- JSON 列
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    data JSONB
);

-- 插入 JSON
INSERT INTO events (data) VALUES ('{"type": "click", "page": "/home"}');

-- 查询 JSON 字段
SELECT * FROM events WHERE data->>'type' = 'click';
SELECT * FROM events WHERE data @> '{"type": "click"}';

-- 更新 JSON
UPDATE events
SET data = data || '{"processed": true}'::jsonb
WHERE id = 1;
```

### 数组操作

```sql
-- 数组列
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    post_id INTEGER,
    names TEXT[]
);

-- 插入数组
INSERT INTO tags (post_id, names) VALUES (1, ARRAY['go', 'backend']);

-- 查询包含某元素
SELECT * FROM tags WHERE 'go' = ANY(names);

-- 展开数组
SELECT post_id, UNNEST(names) as tag FROM tags;
```

## 事务

```sql
-- 开始事务
BEGIN;

-- 或
START TRANSACTION;

-- 执行操作
INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com');
INSERT INTO posts (title, author_id) VALUES ('Hello', 1);

-- 提交
COMMIT;

-- 或回滚
ROLLBACK;

-- 保存点
BEGIN;
INSERT INTO users (name) VALUES ('Alice');
SAVEPOINT sp1;
INSERT INTO users (name) VALUES ('Bob');
-- 回滚到保存点
ROLLBACK TO SAVEPOINT sp1;
COMMIT;  -- 只有 Alice 被提交
```

## 索引策略

### 索引类型

```sql
-- B-tree（默认，适用于大多数场景）
CREATE INDEX idx_users_name ON users(name);

-- 唯一索引
CREATE UNIQUE INDEX idx_users_email ON users(email);

-- 复合索引
CREATE INDEX idx_posts_author_date ON posts(author_id, created_at DESC);

-- 部分索引
CREATE INDEX idx_active_users ON users(email) WHERE status = 'active';

-- 表达式索引
CREATE INDEX idx_users_lower_email ON users(LOWER(email));
```

### 索引建议

```
├── WHERE 条件列 → 建立索引
├── JOIN 条件列 → 建立索引
├── ORDER BY 列 → 建立索引
├── 高选择性列 → 适合索引（如 email）
├── 低选择性列 → 不适合单独索引（如 gender）
├── 频繁更新列 → 谨慎索引（增加写入开销）
└── 查询分析 → 使用 EXPLAIN ANALYZE
```

### 查询分析

```sql
-- 分析查询计划
EXPLAIN SELECT * FROM users WHERE email = 'alice@example.com';

-- 详细分析（包含实际执行时间）
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'alice@example.com';

-- 格式化输出
EXPLAIN (FORMAT JSON) SELECT * FROM users WHERE id = 1;
```

## 性能优化

### 查询优化

```sql
-- ❌ 避免 SELECT *
SELECT * FROM users;

-- ✅ 只选择需要的列
SELECT id, name, email FROM users;

-- ❌ 避免在 WHERE 中使用函数
SELECT * FROM users WHERE YEAR(created_at) = 2024;

-- ✅ 使用范围查询
SELECT * FROM users
WHERE created_at >= '2024-01-01' AND created_at < '2025-01-01';

-- ❌ 避免 LIKE '%keyword%'
SELECT * FROM users WHERE name LIKE '%alice%';

-- ✅ 使用全文搜索或前缀匹配
SELECT * FROM users WHERE name LIKE 'alice%';
```

### 批量操作

```sql
-- ❌ 多次单行插入
INSERT INTO users (name) VALUES ('A');
INSERT INTO users (name) VALUES ('B');
INSERT INTO users (name) VALUES ('C');

-- ✅ 批量插入
INSERT INTO users (name) VALUES ('A'), ('B'), ('C');

-- ✅ 使用 COPY（PostgreSQL，最快）
COPY users (name, email) FROM '/path/to/data.csv' CSV;
```

## SQL 与 Go 类型映射

| SQL 类型 | Go 类型 | 说明 |
|----------|---------|------|
| INTEGER, INT | int, int32, int64 | 整数 |
| BIGINT | int64 | 大整数 |
| REAL, FLOAT | float32 | 单精度浮点 |
| DOUBLE PRECISION | float64 | 双精度浮点 |
| VARCHAR, TEXT | string | 字符串 |
| BOOLEAN | bool | 布尔值 |
| TIMESTAMP | time.Time | 时间戳 |
| DATE | time.Time | 日期 |
| BYTEA | []byte | 二进制 |
| JSONB | json.RawMessage / map | JSON |
| NULL | sql.NullString, *string | 可空值 |

**下一节**：[PostgreSQL 安装](./02-postgres-setup.md) - 学习 PostgreSQL 安装和配置
