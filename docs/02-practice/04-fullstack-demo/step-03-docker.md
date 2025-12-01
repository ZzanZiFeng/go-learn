# Step 3: Docker 部署

## 目标

使用 Docker Compose 部署完整应用。

## 3.1 Docker Compose 配置

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  # PostgreSQL 数据库
  postgres:
    image: postgres:15-alpine
    container_name: fullstack-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sh:/docker-entrypoint-initdb.d/init-db.sh
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - backend

  # Redis 缓存
  redis:
    image: redis:7-alpine
    container_name: fullstack-redis
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - backend

  # Auth Service
  auth-service:
    build:
      context: ../auth-service
      dockerfile: Dockerfile
    container_name: fullstack-auth
    environment:
      - SERVER_PORT=8081
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=auth_service
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_ACCESS_SECRET=${JWT_ACCESS_SECRET}
      - JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - backend

  # Todo API
  todo-api:
    build:
      context: ../todo-api
      dockerfile: Dockerfile
    container_name: fullstack-todo
    environment:
      - SERVER_PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=todo_api
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=${JWT_ACCESS_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      auth-service:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - backend

  # Nginx 网关
  nginx:
    image: nginx:alpine
    container_name: fullstack-nginx
    volumes:
      - ./gateway/nginx.conf:/etc/nginx/conf.d/default.conf:ro
      - ./frontend:/usr/share/nginx/html:ro
    ports:
      - "80:80"
    depends_on:
      - auth-service
      - todo-api
    networks:
      - backend

volumes:
  postgres_data:
  redis_data:

networks:
  backend:
    driver: bridge
```

## 3.2 服务 Dockerfile

创建 `projects/auth-service/Dockerfile`:

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# 复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8081

# 运行
CMD ["./main"]
```

创建 `projects/todo-api/Dockerfile`:

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./main"]
```

## 3.3 数据库初始化脚本

创建 `scripts/init-db.sh`:

```bash
#!/bin/bash
set -e

# 创建数据库
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE auth_service;
    CREATE DATABASE todo_api;
EOSQL

echo "Databases created successfully!"
```

## 3.4 部署脚本

创建 `scripts/deploy.sh`:

```bash
#!/bin/bash

set -e

echo "Starting deployment..."

# 检查 .env 文件
if [ ! -f .env ]; then
    echo "Error: .env file not found"
    echo "Please copy .env.example to .env and configure"
    exit 1
fi

# 加载环境变量
source .env

# 构建镜像
echo "Building images..."
docker-compose build

# 停止旧容器
echo "Stopping old containers..."
docker-compose down

# 启动新容器
echo "Starting new containers..."
docker-compose up -d

# 等待服务就绪
echo "Waiting for services to be ready..."
sleep 10

# 健康检查
echo "Checking health..."
curl -f http://localhost/health/auth || echo "Auth service not ready"
curl -f http://localhost/health/todo || echo "Todo API not ready"

echo "Deployment completed!"
echo "Application is running at http://localhost"
```

## 3.5 常用命令

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f auth-service

# 重启服务
docker-compose restart auth-service

# 停止所有服务
docker-compose down

# 清理数据
docker-compose down -v

# 重新构建
docker-compose up -d --build

# 查看容器状态
docker-compose ps

# 进入容器
docker-compose exec auth-service sh
```

## 3.6 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                     Docker Compose Network                       │
└─────────────────────────────────────────────────────────────────┘

    ┌───────────────────────────────────────────────────────────┐
    │                      backend network                       │
    │  ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐   │
    │  │  nginx  │   │  auth   │   │  todo   │   │postgres │   │
    │  │  :80    │──▶│ :8081   │──▶│ :8080   │──▶│ :5432   │   │
    │  └─────────┘   └─────────┘   └─────────┘   └─────────┘   │
    │       │              │             │             │        │
    │       │              └─────────────┴─────────────┘        │
    │       │                      │                            │
    │       │               ┌──────▼──────┐                     │
    │       │               │    redis    │                     │
    │       │               │   :6379     │                     │
    │       │               └─────────────┘                     │
    └───────┼───────────────────────────────────────────────────┘
            │
    ┌───────▼───────┐
    │   Host :80    │
    └───────────────┘
```

## 检查点

完成本步骤后：

- [x] 创建 Docker Compose 配置
- [x] 编写服务 Dockerfile
- [x] 配置数据库初始化
- [x] 编写部署脚本

## 下一步

[Step 4: 前端集成](./step-04-frontend.md) - 简单前端页面。
