// Package main 演示使用 OpenTelemetry 进行分布式追踪
package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// ==================== 追踪初始化 ====================

// InitTracer 初始化追踪器（使用 stdout exporter 用于演示）
func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
	// 创建 stdout exporter（生产环境应使用 Jaeger 或 OTLP exporter）
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	// 创建资源
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "development"),
		),
	)
	if err != nil {
		return nil, err
	}

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// 设置全局 TracerProvider 和 Propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// ==================== 追踪包装器 ====================

var tracer = otel.Tracer("order-service")

// TracedDB 模拟带追踪的数据库
type TracedDB struct{}

func (db *TracedDB) Query(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	ctx, span := tracer.Start(ctx, "db.query",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		semconv.DBSystemPostgreSQL,
		semconv.DBStatement(query),
		semconv.DBOperation("SELECT"),
	)

	// 模拟数据库查询
	time.Sleep(time.Duration(10+rand.Intn(40)) * time.Millisecond)

	span.SetStatus(codes.Ok, "")
	return nil, nil
}

func (db *TracedDB) Exec(ctx context.Context, query string, args ...interface{}) error {
	ctx, span := tracer.Start(ctx, "db.exec",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		semconv.DBSystemPostgreSQL,
		semconv.DBStatement(query),
	)

	// 模拟数据库执行
	time.Sleep(time.Duration(5+rand.Intn(20)) * time.Millisecond)

	span.SetStatus(codes.Ok, "")
	return nil
}

// TracedCache 模拟带追踪的缓存
type TracedCache struct{}

func (c *TracedCache) Get(ctx context.Context, key string) (string, bool) {
	ctx, span := tracer.Start(ctx, "cache.get",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.system", "redis"),
		attribute.String("cache.key", key),
	)

	// 模拟缓存查询
	time.Sleep(time.Duration(1+rand.Intn(5)) * time.Millisecond)

	// 模拟缓存命中/未命中
	hit := rand.Float32() > 0.3
	if hit {
		span.AddEvent("cache_hit")
		span.SetStatus(codes.Ok, "")
		return "cached_value", true
	}

	span.AddEvent("cache_miss")
	return "", false
}

func (c *TracedCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	ctx, span := tracer.Start(ctx, "cache.set",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.system", "redis"),
		attribute.String("cache.key", key),
		attribute.Int64("cache.ttl_seconds", int64(ttl.Seconds())),
	)

	// 模拟缓存写入
	time.Sleep(time.Duration(1+rand.Intn(3)) * time.Millisecond)

	span.SetStatus(codes.Ok, "")
	return nil
}

// TracedHTTPClient 带追踪的 HTTP 客户端
type TracedHTTPClient struct {
	client *http.Client
}

func NewTracedHTTPClient() *TracedHTTPClient {
	return &TracedHTTPClient{
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (c *TracedHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

// ==================== 业务服务 ====================

// Order 订单
type Order struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	ProductID string  `json:"product_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

// User 用户
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// OrderService 订单服务
type OrderService struct {
	db         *TracedDB
	cache      *TracedCache
	httpClient *TracedHTTPClient
}

func NewOrderService() *OrderService {
	return &OrderService{
		db:         &TracedDB{},
		cache:      &TracedCache{},
		httpClient: NewTracedHTTPClient(),
	}
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	ctx, span := tracer.Start(ctx, "OrderService.GetOrder")
	defer span.End()

	span.SetAttributes(attribute.String("order.id", orderID))

	// 先查缓存
	if _, hit := s.cache.Get(ctx, "order:"+orderID); hit {
		span.AddEvent("order_from_cache")
		return &Order{
			ID:     orderID,
			UserID: "user-123",
			Amount: 99.99,
			Status: "completed",
		}, nil
	}

	// 查数据库
	span.AddEvent("fetching_from_database")
	_, err := s.db.Query(ctx, "SELECT * FROM orders WHERE id = $1", orderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	order := &Order{
		ID:        orderID,
		UserID:    "user-123",
		ProductID: "product-456",
		Amount:    99.99,
		Status:    "completed",
	}

	// 写入缓存
	s.cache.Set(ctx, "order:"+orderID, "order_data", 5*time.Minute)

	span.SetStatus(codes.Ok, "")
	return order, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, productID string, amount float64) (*Order, error) {
	ctx, span := tracer.Start(ctx, "OrderService.CreateOrder")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("product.id", productID),
		attribute.Float64("order.amount", amount),
	)

	// 验证用户
	if err := s.validateUser(ctx, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user validation failed")
		return nil, err
	}

	// 检查库存
	if err := s.checkInventory(ctx, productID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "inventory check failed")
		return nil, err
	}

	// 创建订单
	orderID := fmt.Sprintf("order-%d", rand.Int())
	span.AddEvent("creating_order", trace.WithAttributes(
		attribute.String("order.id", orderID),
	))

	err := s.db.Exec(ctx, "INSERT INTO orders (id, user_id, product_id, amount, status) VALUES ($1, $2, $3, $4, $5)",
		orderID, userID, productID, amount, "pending")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	order := &Order{
		ID:        orderID,
		UserID:    userID,
		ProductID: productID,
		Amount:    amount,
		Status:    "pending",
	}

	span.SetAttributes(attribute.String("order.id", orderID))
	span.SetStatus(codes.Ok, "")
	return order, nil
}

func (s *OrderService) validateUser(ctx context.Context, userID string) error {
	ctx, span := tracer.Start(ctx, "OrderService.validateUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID))

	// 模拟用户服务调用
	// 在生产环境，这里会调用外部服务
	time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)

	span.SetStatus(codes.Ok, "")
	return nil
}

func (s *OrderService) checkInventory(ctx context.Context, productID string) error {
	ctx, span := tracer.Start(ctx, "OrderService.checkInventory")
	defer span.End()

	span.SetAttributes(attribute.String("product.id", productID))

	// 查缓存
	if _, hit := s.cache.Get(ctx, "inventory:"+productID); hit {
		span.AddEvent("inventory_from_cache")
		span.SetStatus(codes.Ok, "")
		return nil
	}

	// 查数据库
	_, err := s.db.Query(ctx, "SELECT quantity FROM inventory WHERE product_id = $1", productID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// 写入缓存
	s.cache.Set(ctx, "inventory:"+productID, "100", time.Minute)

	span.SetStatus(codes.Ok, "")
	return nil
}

// ==================== HTTP 处理器 ====================

func getOrderHandler(svc *OrderService) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		ctx := c.Request.Context()

		order, err := svc.GetOrder(ctx, orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

func createOrderHandler(svc *OrderService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID    string  `json:"user_id"`
			ProductID string  `json:"product_id"`
			Amount    float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()
		order, err := svc.CreateOrder(ctx, req.UserID, req.ProductID, req.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, order)
	}
}

// ==================== 主程序 ====================

func main() {
	// 初始化追踪器
	tp, err := InitTracer("order-service")
	if err != nil {
		panic(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tp.Shutdown(ctx)
	}()

	// 创建服务
	orderService := NewOrderService()

	// 创建 Gin 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 添加 OpenTelemetry 中间件
	r.Use(otelgin.Middleware("order-service"))

	// 路由
	r.GET("/api/orders/:id", getOrderHandler(orderService))
	r.POST("/api/orders", createOrderHandler(orderService))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	fmt.Println("Order Service starting on :8080")
	fmt.Println("Trace output will be printed to stdout")
	fmt.Println("")
	fmt.Println("Test with:")
	fmt.Println("  curl http://localhost:8080/api/orders/123")
	fmt.Println("  curl -X POST http://localhost:8080/api/orders -H 'Content-Type: application/json' -d '{\"user_id\":\"user-1\",\"product_id\":\"product-1\",\"amount\":99.99}'")

	r.Run(":8080")
}

/*
运行示例:
  go run main.go

测试:
  # 获取订单
  curl http://localhost:8080/api/orders/123

  # 创建订单
  curl -X POST http://localhost:8080/api/orders \
    -H "Content-Type: application/json" \
    -d '{"user_id":"user-1","product_id":"product-1","amount":99.99}'

预期追踪输出（简化）:
  Trace ID: abc123...
  ├── Span: HTTP GET /api/orders/:id
  │   └── Span: OrderService.GetOrder
  │       ├── Span: cache.get (cache_miss)
  │       ├── Span: db.query
  │       └── Span: cache.set

  Trace ID: def456...
  ├── Span: HTTP POST /api/orders
  │   └── Span: OrderService.CreateOrder
  │       ├── Span: OrderService.validateUser
  │       ├── Span: OrderService.checkInventory
  │       │   ├── Span: cache.get
  │       │   ├── Span: db.query
  │       │   └── Span: cache.set
  │       └── Span: db.exec

生产环境使用 Jaeger:
  1. 安装 Jaeger exporter:
     go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc

  2. 修改 InitTracer 使用 Jaeger:
     exporter, err := otlptracegrpc.New(ctx,
         otlptracegrpc.WithEndpoint("localhost:4317"),
         otlptracegrpc.WithInsecure(),
     )

  3. 启动 Jaeger:
     docker run -d --name jaeger \
       -p 16686:16686 \
       -p 4317:4317 \
       jaegertracing/all-in-one:1.50

  4. 访问 Jaeger UI: http://localhost:16686
*/
