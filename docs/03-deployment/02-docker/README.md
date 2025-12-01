# Docker 容器化

## 学习目标

掌握 Go 应用的 Docker 容器化部署。

## 1. Dockerfile 基础

### 1.1 简单 Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.21-alpine

WORKDIR /app

# 复制源码
COPY . .

# 下载依赖
RUN go mod download

# 编译
RUN go build -o main ./cmd/api

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./main"]
```

### 1.2 构建和运行

```bash
# 构建镜像
docker build -t myapp:v1 .

# 运行容器
docker run -p 8080:8080 myapp:v1

# 后台运行
docker run -d -p 8080:8080 --name myapp myapp:v1

# 查看日志
docker logs -f myapp

# 停止容器
docker stop myapp
```

## 2. 多阶段构建

### 2.1 优化的 Dockerfile

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w" \
    -o main ./cmd/api

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates（用于 HTTPS）
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# 运行
CMD ["./main"]
```

### 2.2 镜像大小对比

| 构建方式 | 镜像大小 |
|----------|----------|
| golang:1.21 | ~1.2GB |
| golang:1.21-alpine | ~400MB |
| 多阶段 (alpine) | ~20MB |
| 多阶段 (scratch) | ~10MB |

### 2.3 使用 scratch 基础镜像

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags '-extldflags "-static" -s -w' \
    -o main ./cmd/api

# 运行阶段 - 最小镜像
FROM scratch

# 复制 CA 证书
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 复制二进制文件
COPY --from=builder /app/main /main

EXPOSE 8080

ENTRYPOINT ["/main"]
```

## 3. Docker Compose

### 3.1 基础配置

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=myapp
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - backend

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: myapp
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - backend

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - backend

volumes:
  postgres_data:
  redis_data:

networks:
  backend:
    driver: bridge
```

### 3.2 开发环境配置

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=debug
    command: go run ./cmd/api
```

### 3.3 Dockerfile.dev

```dockerfile
# Dockerfile.dev - 开发环境
FROM golang:1.21

WORKDIR /app

# 安装热重载工具
RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
```

### 3.4 Air 配置

```toml
# .air.toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/api"
bin = "./tmp/main"
full_bin = "./tmp/main"
include_ext = ["go", "tpl", "tmpl", "html", "yaml"]
exclude_dir = ["assets", "tmp", "vendor"]
delay = 1000

[log]
time = false

[color]
main = "magenta"
watcher = "cyan"
build = "yellow"
runner = "green"
```

## 4. Docker 命令速查

```bash
# 构建镜像
docker build -t myapp:v1 .
docker build -t myapp:v1 -f Dockerfile.prod .

# 运行容器
docker run -d -p 8080:8080 --name myapp myapp:v1
docker run -it --rm myapp:v1 sh  # 交互模式

# 查看容器
docker ps
docker ps -a

# 日志
docker logs myapp
docker logs -f --tail 100 myapp

# 进入容器
docker exec -it myapp sh

# 停止/删除
docker stop myapp
docker rm myapp
docker rmi myapp:v1

# Docker Compose
docker-compose up -d
docker-compose down
docker-compose logs -f
docker-compose ps
docker-compose exec app sh

# 清理
docker system prune -a  # 清理未使用的镜像和容器
```

## 5. 最佳实践

### 5.1 .dockerignore

```
# .dockerignore
.git
.gitignore
.env
.env.*
*.md
Makefile
docker-compose*.yml
Dockerfile*
.dockerignore
tmp/
bin/
*.test
*.out
.idea/
.vscode/
```

### 5.2 安全实践

```dockerfile
# 1. 使用非 root 用户
FROM alpine:latest

RUN adduser -D -g '' appuser
USER appuser

COPY --from=builder --chown=appuser:appuser /app/main /main

# 2. 只复制必要文件
COPY --from=builder /app/main /main
COPY --from=builder /app/configs /configs

# 3. 设置只读文件系统
# docker run --read-only myapp:v1
```

## 练习

1. 为你的项目创建多阶段 Dockerfile
2. 配置 Docker Compose 开发环境
3. 添加健康检查
4. 比较不同基础镜像的大小

## 下一步

[配置管理](../03-configuration/) - 管理应用配置和环境变量。
