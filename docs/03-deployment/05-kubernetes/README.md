# Kubernetes 部署

## 学习目标

掌握 Go 应用在 Kubernetes 上的部署和运维基础。

## 1. Kubernetes 基础概念

### 1.1 核心组件

| 组件 | 说明 |
|------|------|
| Pod | 最小部署单元，一个或多个容器 |
| Deployment | 管理 Pod 副本和更新策略 |
| Service | 提供稳定的访问入口 |
| ConfigMap | 存储配置数据 |
| Secret | 存储敏感数据 |
| Ingress | HTTP 路由规则 |

### 1.2 本地开发环境

```bash
# 安装 minikube (macOS)
brew install minikube

# 启动集群
minikube start

# 查看状态
minikube status

# 启用 Ingress
minikube addons enable ingress

# 打开 Dashboard
minikube dashboard
```

## 2. 部署配置

### 2.1 Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  labels:
    app: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
        - name: myapp
          image: myapp:v1
          ports:
            - containerPort: 8080
          env:
            - name: DB_HOST
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: db-host
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: db-password
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 15
            periodSeconds: 20
```

### 2.2 Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  selector:
    app: myapp
  ports:
    - port: 80
      targetPort: 8080
  type: ClusterIP

---
# NodePort 类型（开发测试）
apiVersion: v1
kind: Service
metadata:
  name: myapp-nodeport
spec:
  selector:
    app: myapp
  ports:
    - port: 80
      targetPort: 8080
      nodePort: 30080
  type: NodePort
```

### 2.3 ConfigMap

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  db-host: "postgres.default.svc.cluster.local"
  db-port: "5432"
  db-name: "myapp"
  redis-host: "redis.default.svc.cluster.local"
  redis-port: "6379"
  log-level: "info"
```

### 2.4 Secret

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
stringData:  # 自动 base64 编码
  db-password: "your-password"
  jwt-secret: "your-jwt-secret"
```

```bash
# 从命令行创建 Secret
kubectl create secret generic app-secrets \
  --from-literal=db-password=your-password \
  --from-literal=jwt-secret=your-jwt-secret
```

### 2.5 Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
    - host: myapp.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: myapp
                port:
                  number: 80
```

## 3. 健康检查端点

### 3.1 Go 实现

```go
// internal/handlers/health.go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type HealthHandler struct {
    db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
    return &HealthHandler{db: db}
}

// Health 健康检查（用于 readiness/liveness probe）
func (h *HealthHandler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
    })
}

// Ready 就绪检查（检查依赖服务）
func (h *HealthHandler) Ready(c *gin.Context) {
    // 检查数据库连接
    sqlDB, err := h.db.DB()
    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "not ready",
            "error":  "database connection failed",
        })
        return
    }

    if err := sqlDB.Ping(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "not ready",
            "error":  "database ping failed",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "ready",
    })
}
```

### 3.2 路由配置

```go
func SetupRoutes(r *gin.Engine, healthHandler *HealthHandler) {
    // 健康检查端点（不需要认证）
    r.GET("/health", healthHandler.Health)
    r.GET("/ready", healthHandler.Ready)
}
```

## 4. kubectl 常用命令

### 4.1 部署管理

```bash
# 应用配置
kubectl apply -f k8s/

# 查看部署状态
kubectl get deployments
kubectl get pods
kubectl get services

# 查看 Pod 详情
kubectl describe pod myapp-xxxxx

# 查看日志
kubectl logs myapp-xxxxx
kubectl logs -f myapp-xxxxx  # 实时跟踪

# 进入 Pod
kubectl exec -it myapp-xxxxx -- sh

# 端口转发（本地调试）
kubectl port-forward pod/myapp-xxxxx 8080:8080
kubectl port-forward svc/myapp 8080:80
```

### 4.2 扩缩容

```bash
# 手动扩容
kubectl scale deployment myapp --replicas=5

# 自动扩缩容
kubectl autoscale deployment myapp --min=2 --max=10 --cpu-percent=80

# 查看 HPA 状态
kubectl get hpa
```

### 4.3 更新和回滚

```bash
# 更新镜像
kubectl set image deployment/myapp myapp=myapp:v2

# 查看更新状态
kubectl rollout status deployment/myapp

# 查看历史
kubectl rollout history deployment/myapp

# 回滚
kubectl rollout undo deployment/myapp
kubectl rollout undo deployment/myapp --to-revision=2
```

## 5. 完整部署示例

### 5.1 目录结构

```
k8s/
├── base/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   └── kustomization.yaml
├── overlays/
│   ├── development/
│   │   ├── kustomization.yaml
│   │   └── patches/
│   └── production/
│       ├── kustomization.yaml
│       └── patches/
└── secrets/
    └── secret.yaml
```

### 5.2 Kustomization

```yaml
# k8s/base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - deployment.yaml
  - service.yaml
  - configmap.yaml
```

```yaml
# k8s/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: production

resources:
  - ../../base

patchesStrategicMerge:
  - patches/deployment-replicas.yaml

images:
  - name: myapp
    newTag: v1.0.0
```

```yaml
# k8s/overlays/production/patches/deployment-replicas.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 5
```

```bash
# 应用生产配置
kubectl apply -k k8s/overlays/production/
```

## 6. 监控和日志

### 6.1 Prometheus ServiceMonitor

```yaml
# k8s/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: myapp
  labels:
    app: myapp
spec:
  selector:
    matchLabels:
      app: myapp
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
```

### 6.2 日志采集配置

```yaml
# Pod 标准输出日志会被自动采集
# 确保应用日志输出到 stdout/stderr
spec:
  containers:
    - name: myapp
      # 日志输出到 stdout
```

## 7. 部署流程图

```
┌──────────────────────────────────────────────────────────────────┐
│                    Kubernetes 部署流程                            │
└──────────────────────────────────────────────────────────────────┘

 开发者                    CI/CD                   Kubernetes
    │                        │                        │
    │ 1. Push Code           │                        │
    │ ──────────────────────▶│                        │
    │                        │                        │
    │                        │ 2. Build & Test        │
    │                        │ ──────────────────────▶│
    │                        │                        │
    │                        │ 3. Push Image          │
    │                        │ ──────────────────────▶│ Registry
    │                        │                        │
    │                        │ 4. kubectl apply       │
    │                        │ ──────────────────────▶│
    │                        │                        │
    │                        │                        │ 5. Create Pods
    │                        │                        │ ─────────────▶
    │                        │                        │
    │                        │                        │ 6. Health Check
    │                        │                        │ ─────────────▶
    │                        │                        │
    │                        │ 7. Rollout Complete    │
    │                        │◀────────────────────── │
    │                        │                        │
```

## 练习

1. 使用 minikube 部署一个 Go 应用
2. 配置 ConfigMap 和 Secret
3. 实现滚动更新和回滚
4. 配置水平自动扩缩容 (HPA)

## 总结

| 主题 | 要点 |
|------|------|
| Deployment | 管理 Pod 副本、更新策略 |
| Service | 提供稳定访问入口 |
| ConfigMap/Secret | 配置和敏感信息管理 |
| Ingress | HTTP 路由 |
| 健康检查 | readiness/liveness probe |
| 监控 | Prometheus + ServiceMonitor |

## 下一步学习

- **[测试轨道](../../04-testing/)** - 单元测试、集成测试
- Helm Charts 进阶
- Service Mesh (Istio)
- GitOps (ArgoCD)
