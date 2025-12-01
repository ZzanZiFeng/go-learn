# 项目 4: Fullstack Demo

## 项目概述

整合 Todo API 和 Auth Service，构建一个完整的全栈演示项目。

| 难度 | 预计时间 | 前置项目 |
|------|----------|----------|
| 高级 | 6-8 小时 | Todo API, Auth Service |

## 学习目标

- 微服务架构设计
- API 网关模式
- 前后端集成
- Docker 部署

## 项目架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        Fullstack Demo                           │
└─────────────────────────────────────────────────────────────────┘

                           ┌─────────────┐
                           │   Nginx     │
                           │  (网关)     │
                           └──────┬──────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
              ▼                   ▼                   ▼
      ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
      │  Frontend     │   │  Auth Service │   │   Todo API    │
      │  (静态文件)   │   │   :8081       │   │    :8080      │
      └───────────────┘   └───────┬───────┘   └───────┬───────┘
                                  │                   │
                          ┌───────┴───────────────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
              ▼           ▼           ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │ PostgreSQL│ │  Redis   │ │ RabbitMQ │
        └──────────┘ └──────────┘ └──────────┘
```

## 技术栈

```yaml
网关: Nginx / Traefik
后端: Go (Gin)
数据库: PostgreSQL
缓存: Redis
消息队列: RabbitMQ (可选)
容器: Docker + Docker Compose
监控: Prometheus + Grafana (可选)
```

## 项目结构

```
projects/fullstack-demo/
├── docker-compose.yml
├── .env.example
├── gateway/
│   └── nginx.conf
├── services/
│   ├── auth/           # Auth Service (复用)
│   └── todo/           # Todo API (复用)
├── frontend/
│   └── index.html
├── scripts/
│   ├── init-db.sh
│   └── deploy.sh
└── README.md
```

## 学习步骤

1. **[Step 1: 项目整合](./step-01-integration.md)** - 整合服务
2. **[Step 2: API 网关](./step-02-gateway.md)** - 配置 Nginx
3. **[Step 3: Docker 部署](./step-03-docker.md)** - 容器化
4. **[Step 4: 前端集成](./step-04-frontend.md)** - 简单前端

## 知识点映射

| 步骤 | 知识点 | 开发轨道章节 |
|------|--------|--------------|
| Step 1 | 服务间通信 | DEV-04 |
| Step 2 | 反向代理、负载均衡 | 部署轨道 |
| Step 3 | Docker、编排 | 部署轨道 |
| Step 4 | RESTful API 集成 | DEV-05 |

## 快速开始

```bash
# 克隆项目
cd projects/fullstack-demo

# 复制环境变量
cp .env.example .env

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 访问应用
open http://localhost
```

## API 端点汇总

### Auth Service (:8081)

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/auth/register | 注册 |
| POST | /api/auth/login | 登录 |
| POST | /api/auth/refresh | 刷新令牌 |
| GET | /api/users/me | 获取当前用户 |

### Todo API (:8080)

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/todos | 获取待办列表 |
| POST | /api/todos | 创建待办 |
| PUT | /api/todos/:id | 更新待办 |
| DELETE | /api/todos/:id | 删除待办 |

## 下一步

完成 Fullstack Demo 后，继续学习：

- **[部署轨道](../../03-deployment/)** - 生产环境部署
- **[测试轨道](../../04-testing/)** - 自动化测试
