# 微服务入门 (Microservices Introduction)

## 什么是微服务？

微服务架构将单体应用拆分为多个小型、独立部署的服务，每个服务专注于特定的业务功能。

### 单体 vs 微服务

```
┌─────────────────────────────────────┐
│           单体应用                   │
│  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐  │
│  │用户 │ │订单 │ │支付 │ │库存 │  │
│  │模块 │ │模块 │ │模块 │ │模块 │  │
│  └─────┘ └─────┘ └─────┘ └─────┘  │
│            共享数据库               │
└─────────────────────────────────────┘

             ↓ 拆分为 ↓

┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐
│用户   │  │订单   │  │支付   │  │库存   │
│服务   │  │服务   │  │服务   │  │服务   │
└──┬───┘  └──┬───┘  └──┬───┘  └──┬───┘
   │         │         │         │
┌──┴──┐   ┌──┴──┐   ┌──┴──┐   ┌──┴──┐
│ DB  │   │ DB  │   │ DB  │   │ DB  │
└─────┘   └─────┘   └─────┘   └─────┘
```

## 与 Node.js 对比

| 概念 | Node.js | Go |
|------|---------|-----|
| HTTP 服务 | Express, Fastify | net/http, Gin |
| gRPC | @grpc/grpc-js | google.golang.org/grpc |
| 服务发现 | Consul, etcd SDK | Consul, etcd 原生支持 |
| 消息队列 | amqplib, kafka-node | amqp091-go, segmentio/kafka-go |

## API 网关模式

API 网关是微服务的入口点，处理：
- 请求路由
- 认证授权
- 限流
- 负载均衡
- 请求聚合

```
                    ┌─────────────────┐
      客户端 ───────→│   API Gateway   │
                    │  (Kong/Nginx)   │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
    ┌─────────┐        ┌─────────┐        ┌─────────┐
    │  用户   │        │  订单   │        │  支付   │
    │  服务   │        │  服务   │        │  服务   │
    └─────────┘        └─────────┘        └─────────┘
```

### 简单 API 网关实现

```go
// cmd/gateway/main.go
package main

import (
    "net/http"
    "net/http/httputil"
    "net/url"

    "github.com/gin-gonic/gin"
)

type ServiceConfig struct {
    Name     string
    URL      string
    PathPrefix string
}

var services = []ServiceConfig{
    {Name: "user", URL: "http://user-service:8081", PathPrefix: "/api/users"},
    {Name: "order", URL: "http://order-service:8082", PathPrefix: "/api/orders"},
    {Name: "product", URL: "http://product-service:8083", PathPrefix: "/api/products"},
}

func main() {
    r := gin.Default()

    // 注册代理
    for _, svc := range services {
        setupProxy(r, svc)
    }

    r.Run(":8080")
}

func setupProxy(r *gin.Engine, svc ServiceConfig) {
    target, _ := url.Parse(svc.URL)
    proxy := httputil.NewSingleHostReverseProxy(target)

    r.Any(svc.PathPrefix+"/*path", func(c *gin.Context) {
        // 可以在这里添加认证、限流等
        proxy.ServeHTTP(c.Writer, c.Request)
    })
}
```

## 服务间通信

### 1. HTTP/REST

最简单，适合同步调用：

```go
// internal/clients/user_client.go
package clients

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type UserClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewUserClient(baseURL string) *UserClient {
    return &UserClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

type User struct {
    ID    uint   `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}

func (c *UserClient) GetUser(ctx context.Context, userID uint) (*User, error) {
    url := fmt.Sprintf("%s/api/users/%d", c.baseURL, userID)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, ErrUserNotFound
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }

    return &user, nil
}
```

### 2. gRPC

高性能，适合内部服务间调用：

```protobuf
// api/proto/user.proto
syntax = "proto3";

package user;
option go_package = "myapp/api/proto/user";

service UserService {
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
}

message GetUserRequest {
  uint64 id = 1;
}

message CreateUserRequest {
  string email = 1;
  string name = 2;
}

message UserResponse {
  uint64 id = 1;
  string email = 2;
  string name = 3;
}
```

```bash
# 生成 Go 代码
protoc --go_out=. --go-grpc_out=. api/proto/user.proto
```

```go
// gRPC Server
package main

import (
    "context"
    "net"

    "google.golang.org/grpc"
    pb "myapp/api/proto/user"
)

type userServer struct {
    pb.UnimplementedUserServiceServer
    service UserService
}

func (s *userServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
    user, err := s.service.GetUserByID(ctx, uint(req.Id))
    if err != nil {
        return nil, err
    }
    return &pb.UserResponse{
        Id:    uint64(user.ID),
        Email: user.Email,
        Name:  user.Name,
    }, nil
}

func main() {
    lis, _ := net.Listen("tcp", ":50051")
    s := grpc.NewServer()
    pb.RegisterUserServiceServer(s, &userServer{})
    s.Serve(lis)
}
```

```go
// gRPC Client
package clients

import (
    "context"

    "google.golang.org/grpc"
    pb "myapp/api/proto/user"
)

type UserGRPCClient struct {
    client pb.UserServiceClient
}

func NewUserGRPCClient(addr string) (*UserGRPCClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return &UserGRPCClient{
        client: pb.NewUserServiceClient(conn),
    }, nil
}

func (c *UserGRPCClient) GetUser(ctx context.Context, id uint) (*User, error) {
    resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{Id: uint64(id)})
    if err != nil {
        return nil, err
    }
    return &User{
        ID:    uint(resp.Id),
        Email: resp.Email,
        Name:  resp.Name,
    }, nil
}
```

### 3. 消息队列（异步）

适合解耦和异步处理：

```go
// 使用 RabbitMQ
package messaging

import (
    "encoding/json"

    amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
    channel *amqp.Channel
}

func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }
    return &Publisher{channel: ch}, nil
}

type UserCreatedEvent struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
}

func (p *Publisher) PublishUserCreated(event UserCreatedEvent) error {
    body, err := json.Marshal(event)
    if err != nil {
        return err
    }

    return p.channel.Publish(
        "user_events", // exchange
        "user.created", // routing key
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        },
    )
}
```

```go
// Consumer
type Consumer struct {
    channel *amqp.Channel
}

func (c *Consumer) ConsumeUserCreated(handler func(UserCreatedEvent) error) error {
    msgs, err := c.channel.Consume(
        "user_created_queue",
        "",
        false, // auto-ack
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return err
    }

    go func() {
        for msg := range msgs {
            var event UserCreatedEvent
            if err := json.Unmarshal(msg.Body, &event); err != nil {
                msg.Nack(false, true)
                continue
            }

            if err := handler(event); err != nil {
                msg.Nack(false, true)
                continue
            }

            msg.Ack(false)
        }
    }()

    return nil
}
```

## 服务发现

### 使用 Consul

```go
package discovery

import (
    "fmt"

    "github.com/hashicorp/consul/api"
)

type ServiceDiscovery struct {
    client *api.Client
}

func NewServiceDiscovery(addr string) (*ServiceDiscovery, error) {
    config := api.DefaultConfig()
    config.Address = addr

    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }

    return &ServiceDiscovery{client: client}, nil
}

// Register 注册服务
func (sd *ServiceDiscovery) Register(name, id, addr string, port int) error {
    return sd.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
        ID:      id,
        Name:    name,
        Address: addr,
        Port:    port,
        Check: &api.AgentServiceCheck{
            HTTP:     fmt.Sprintf("http://%s:%d/health", addr, port),
            Interval: "10s",
            Timeout:  "5s",
        },
    })
}

// Discover 发现服务
func (sd *ServiceDiscovery) Discover(name string) ([]*api.ServiceEntry, error) {
    entries, _, err := sd.client.Health().Service(name, "", true, nil)
    return entries, err
}

// Deregister 注销服务
func (sd *ServiceDiscovery) Deregister(id string) error {
    return sd.client.Agent().ServiceDeregister(id)
}
```

## 分布式追踪

### 使用 OpenTelemetry

```go
package tracing

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracer(serviceName, jaegerEndpoint string) (*tracesdk.TracerProvider, error) {
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
    if err != nil {
        return nil, err
    }

    tp := tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(exp),
        tracesdk.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}

// 使用
func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
    tracer := otel.Tracer("user-service")
    ctx, span := tracer.Start(ctx, "GetUser")
    defer span.End()

    // 添加属性
    span.SetAttributes(attribute.Int("user.id", int(id)))

    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    return user, nil
}
```

## 微服务设计原则

### 1. 单一职责

每个服务只负责一个业务领域：
- 用户服务：用户注册、认证、资料管理
- 订单服务：订单创建、状态管理
- 支付服务：支付处理、退款

### 2. 数据独立

每个服务拥有自己的数据库：
```
# ✗ 共享数据库
user-service ──→┐
                ├──→ shared-database
order-service ─→┘

# ✓ 独立数据库
user-service  ──→ user-db
order-service ──→ order-db
```

### 3. API 优先

先定义 API 契约：
```yaml
# api/openapi/user-service.yaml
openapi: 3.0.0
info:
  title: User Service API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      summary: Get user by ID
      responses:
        200:
          description: User found
```

### 4. 容错设计

- **超时**：设置合理的超时时间
- **重试**：实现指数退避重试
- **熔断**：使用熔断器防止级联故障
- **降级**：服务不可用时提供降级响应

```go
// 使用 hystrix-go 实现熔断
import "github.com/afex/hystrix-go/hystrix"

func init() {
    hystrix.ConfigureCommand("get_user", hystrix.CommandConfig{
        Timeout:               1000,
        MaxConcurrentRequests: 100,
        ErrorPercentThreshold: 50,
    })
}

func (c *UserClient) GetUserWithCircuitBreaker(ctx context.Context, id uint) (*User, error) {
    var user *User

    err := hystrix.Do("get_user", func() error {
        var err error
        user, err = c.GetUser(ctx, id)
        return err
    }, func(err error) error {
        // 降级逻辑
        return ErrServiceUnavailable
    })

    return user, err
}
```

## 何时使用微服务

### 适合微服务

- 团队规模大（多个小团队）
- 需要独立部署不同模块
- 不同模块有不同的扩展需求
- 技术栈多样化需求

### 不适合微服务

- 小型项目/团队
- 业务边界不清晰
- 没有 DevOps 能力
- 对延迟要求极高

## 总结

| 模式 | 用途 |
|------|------|
| API 网关 | 统一入口，认证、限流 |
| HTTP/REST | 简单的同步调用 |
| gRPC | 高性能内部通信 |
| 消息队列 | 异步解耦 |
| 服务发现 | 动态服务定位 |
| 分布式追踪 | 请求链路追踪 |
| 熔断器 | 容错和降级 |

微服务带来的复杂性：
- 分布式事务
- 服务间通信
- 数据一致性
- 部署和运维复杂度

**建议**：从单体开始，在需要时拆分为微服务。
