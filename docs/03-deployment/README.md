# 部署教程概览

## 学习目标

掌握 Go 应用的编译、打包、容器化和生产部署。

## 章节结构

| 章节 | 主题 | 难度 | 时间 |
|------|------|------|------|
| DEPLOY-01 | 编译与打包 | 入门 | 1小时 |
| DEPLOY-02 | Docker 容器化 | 中级 | 2小时 |
| DEPLOY-03 | 配置管理 | 中级 | 1小时 |
| DEPLOY-04 | CI/CD 流水线 | 中级 | 2小时 |
| DEPLOY-05 | Kubernetes 部署 | 高级 | 3小时 |

## 学习路径

```
┌─────────────────────────────────────────────────────────────────┐
│                       部署轨道学习路径                           │
└─────────────────────────────────────────────────────────────────┘

 DEPLOY-01          DEPLOY-02          DEPLOY-03
┌──────────┐      ┌──────────┐      ┌──────────┐
│ 编译打包  │─────▶│  Docker  │─────▶│ 配置管理  │
│ go build │      │ 容器化    │      │ 环境变量  │
└──────────┘      └──────────┘      └──────────┘
                        │
                        ▼
              ┌──────────┐      ┌──────────┐
              │  CI/CD   │─────▶│   K8s    │
              │ 自动化    │      │  部署    │
              └──────────┘      └──────────┘
               DEPLOY-04         DEPLOY-05
```

## 前置要求

- 完成开发轨道 DEV-04（项目结构）
- 安装 Docker
- 了解基本的 Linux 命令

## 目录

1. **[编译与打包](./01-compilation/)** - go build、交叉编译
2. **[Docker 容器化](./02-docker/)** - Dockerfile、Docker Compose
3. **[配置管理](./03-configuration/)** - 环境变量、配置文件
4. **[CI/CD 流水线](./04-cicd/)** - GitHub Actions、自动化部署
5. **[Kubernetes 部署](./05-kubernetes/)** - K8s 基础、Deployment、Service

## 工具清单

| 工具 | 用途 | 安装方式 |
|------|------|----------|
| Docker | 容器化 | brew install docker |
| docker-compose | 编排 | brew install docker-compose |
| kubectl | K8s 命令行 | brew install kubectl |
| minikube | 本地 K8s | brew install minikube |
| GitHub CLI | CI/CD | brew install gh |

## 下一步

从 **[编译与打包](./01-compilation/)** 开始学习！
