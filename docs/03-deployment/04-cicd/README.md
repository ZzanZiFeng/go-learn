# CI/CD 流水线

## 学习目标

掌握 Go 项目的持续集成和持续部署。

## 1. GitHub Actions 基础

### 1.1 工作流文件结构

```yaml
# .github/workflows/ci.yml
name: CI                    # 工作流名称

on:                         # 触发条件
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:                       # 任务定义
  build:
    runs-on: ubuntu-latest  # 运行环境
    steps:                  # 步骤
      - name: Checkout
        uses: actions/checkout@v4
```

### 1.2 基本 Go CI 工作流

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.21'

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: coverage.out

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Build
        run: |
          CGO_ENABLED=0 GOOS=linux go build \
            -ldflags "-s -w -X main.Version=${{ github.sha }}" \
            -o bin/app ./cmd/api

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: app
          path: bin/app
```

## 2. Docker 镜像构建

### 2.1 构建并推送镜像

```yaml
# .github/workflows/docker.yml
name: Docker Build

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### 2.2 多平台构建

```yaml
- name: Set up QEMU
  uses: docker/setup-qemu-action@v3

- name: Build and push (multi-platform)
  uses: docker/build-push-action@v5
  with:
    context: .
    platforms: linux/amd64,linux/arm64
    push: true
    tags: ${{ steps.meta.outputs.tags }}
```

## 3. 集成测试

### 3.1 带服务的测试

```yaml
# .github/workflows/integration.yml
name: Integration Tests

on:
  push:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run integration tests
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: test
          DB_PASSWORD: test
          DB_NAME: testdb
          REDIS_HOST: localhost
          REDIS_PORT: 6379
        run: |
          go test -v -tags=integration ./tests/integration/...
```

## 4. 自动部署

### 4.1 部署到服务器

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    tags: [ 'v*' ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to server
        uses: appleboy/ssh-action@v1.0.0
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            cd /opt/myapp
            docker-compose pull
            docker-compose up -d
            docker image prune -f
```

### 4.2 部署到 Kubernetes

```yaml
# .github/workflows/k8s-deploy.yml
name: K8s Deploy

on:
  push:
    tags: [ 'v*' ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up kubectl
        uses: azure/setup-kubectl@v3

      - name: Configure kubeconfig
        run: |
          mkdir -p ~/.kube
          echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > ~/.kube/config

      - name: Update image tag
        run: |
          TAG=${GITHUB_REF#refs/tags/}
          kubectl set image deployment/myapp \
            myapp=ghcr.io/${{ github.repository }}:$TAG \
            -n production

      - name: Wait for rollout
        run: |
          kubectl rollout status deployment/myapp -n production --timeout=300s
```

## 5. 发布管理

### 5.1 自动创建 Release

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags: [ 'v*' ]

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 5.2 GoReleaser 配置

```yaml
# .goreleaser.yml
project_name: myapp

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.Version={{.Version}}
      - -X main.Commit={{.Commit}}
      - -X main.Date={{.Date}}
    main: ./cmd/api

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'
```

## 6. 工作流最佳实践

### 6.1 使用矩阵构建

```yaml
jobs:
  test:
    strategy:
      matrix:
        go-version: ['1.20', '1.21', '1.22']
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - run: go test ./...
```

### 6.2 缓存依赖

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.21'
    cache: true  # 自动缓存 Go 模块

# 或手动配置
- name: Cache Go modules
  uses: actions/cache@v3
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

### 6.3 使用复合动作

```yaml
# .github/actions/setup/action.yml
name: 'Setup Go Environment'
description: 'Setup Go with caching'

inputs:
  go-version:
    description: 'Go version'
    required: false
    default: '1.21'

runs:
  using: 'composite'
  steps:
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ inputs.go-version }}
        cache: true

    - name: Download dependencies
      shell: bash
      run: go mod download
```

```yaml
# 使用复合动作
steps:
  - uses: actions/checkout@v4
  - uses: ./.github/actions/setup
    with:
      go-version: '1.21'
```

## 7. Secrets 管理

### 7.1 配置 Secrets

```bash
# GitHub CLI 设置 secrets
gh secret set DB_PASSWORD
gh secret set JWT_SECRET
gh secret set SSH_PRIVATE_KEY < ~/.ssh/deploy_key
```

### 7.2 在工作流中使用

```yaml
env:
  DB_PASSWORD: ${{ secrets.DB_PASSWORD }}

steps:
  - name: Deploy
    env:
      SSH_KEY: ${{ secrets.SSH_PRIVATE_KEY }}
    run: |
      echo "$SSH_KEY" > deploy_key
      chmod 600 deploy_key
      ssh -i deploy_key user@server "deploy.sh"
```

## 完整 CI/CD 流程

```
┌──────────────────────────────────────────────────────────────────┐
│                      CI/CD Pipeline                               │
└──────────────────────────────────────────────────────────────────┘

 Push/PR                  Tests Pass               Tag Created
    │                         │                         │
    ▼                         ▼                         ▼
┌────────┐              ┌────────┐              ┌────────┐
│  Lint  │─────────────▶│  Test  │─────────────▶│ Build  │
└────────┘              └────────┘              └────────┘
                                                     │
                                                     ▼
                                               ┌────────┐
                                               │ Docker │
                                               │  Push  │
                                               └────────┘
                                                     │
                                                     ▼
                                               ┌────────┐
                                               │ Deploy │
                                               │  K8s   │
                                               └────────┘
```

## 练习

1. 为你的项目创建完整的 CI 工作流
2. 配置 Docker 镜像自动构建
3. 实现自动部署到测试环境
4. 使用 GoReleaser 自动发布

## 下一步

[运维监控](../05-operations/) - 生产环境监控和告警。
