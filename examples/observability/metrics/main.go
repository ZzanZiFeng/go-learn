// Package main 演示使用 Prometheus 收集指标
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ==================== 指标定义 ====================

const namespace = "myapp"

var (
	// Counter: HTTP 请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Histogram: HTTP 请求延迟
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// Histogram: HTTP 请求大小
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// Histogram: HTTP 响应大小
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// Gauge: 活跃连接数
	activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "active_connections",
		Help:      "Number of active connections",
	})

	// Counter: 业务指标 - 订单创建
	ordersCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "orders",
			Name:      "created_total",
			Help:      "Total number of orders created",
		},
		[]string{"status"},
	)

	// Histogram: 业务指标 - 订单金额
	orderAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "orders",
			Name:      "amount_dollars",
			Help:      "Order amount distribution in dollars",
			Buckets:   []float64{10, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		},
		[]string{"category"},
	)

	// Counter: 业务指标 - 支付
	paymentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "payments",
			Name:      "total",
			Help:      "Total number of payments",
		},
		[]string{"method", "status"},
	)

	// Gauge: 业务指标 - 库存
	inventoryLevel = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "inventory",
			Name:      "level",
			Help:      "Current inventory level",
		},
		[]string{"product_id"},
	)

	// Counter: 缓存操作
	cacheOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "cache",
			Name:      "operations_total",
			Help:      "Total cache operations",
		},
		[]string{"operation", "result"},
	)

	// Histogram: 数据库查询延迟
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "db",
			Name:      "query_duration_seconds",
			Help:      "Database query duration",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
		},
		[]string{"operation", "table"},
	)
)

// ==================== Prometheus 中间件 ====================

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取路由模式（避免高基数）
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// 增加活跃连接数
		activeConnections.Inc()
		defer activeConnections.Dec()

		// 记录请求大小
		reqSize := float64(c.Request.ContentLength)
		if reqSize > 0 {
			httpRequestSize.WithLabelValues(c.Request.Method, path).Observe(reqSize)
		}

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		respSize := float64(c.Writer.Size())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
		httpResponseSize.WithLabelValues(c.Request.Method, path).Observe(respSize)
	}
}

// ==================== 业务逻辑模拟 ====================

// Order 订单
type Order struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Status   string  `json:"status"`
}

// 模拟创建订单
func createOrder(category string, amount float64) (*Order, error) {
	// 模拟数据库操作
	timer := prometheus.NewTimer(dbQueryDuration.WithLabelValues("insert", "orders"))
	time.Sleep(time.Duration(10+rand.Intn(40)) * time.Millisecond)
	timer.ObserveDuration()

	// 模拟缓存操作
	if rand.Float32() > 0.3 {
		cacheOperationsTotal.WithLabelValues("get", "hit").Inc()
	} else {
		cacheOperationsTotal.WithLabelValues("get", "miss").Inc()
	}

	// 记录业务指标
	ordersCreatedTotal.WithLabelValues("success").Inc()
	orderAmount.WithLabelValues(category).Observe(amount)

	return &Order{
		ID:       fmt.Sprintf("order-%d", rand.Int()),
		UserID:   "user-123",
		Amount:   amount,
		Category: category,
		Status:   "created",
	}, nil
}

// 模拟支付
func processPayment(method string, amount float64) error {
	// 模拟外部 API 调用
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	// 模拟支付结果
	success := rand.Float32() > 0.1
	if success {
		paymentsTotal.WithLabelValues(method, "success").Inc()
		return nil
	}

	paymentsTotal.WithLabelValues(method, "failed").Inc()
	return fmt.Errorf("payment failed")
}

// 模拟获取库存
func getInventory(productID string) int {
	// 模拟数据库查询
	timer := prometheus.NewTimer(dbQueryDuration.WithLabelValues("select", "inventory"))
	time.Sleep(time.Duration(5+rand.Intn(20)) * time.Millisecond)
	timer.ObserveDuration()

	level := rand.Intn(100)
	inventoryLevel.WithLabelValues(productID).Set(float64(level))
	return level
}

// ==================== HTTP 处理器 ====================

func createOrderHandler(c *gin.Context) {
	var req struct {
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ordersCreatedTotal.WithLabelValues("failed").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := createOrder(req.Category, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func processPaymentHandler(c *gin.Context) {
	var req struct {
		Method string  `json:"method"`
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := processPayment(req.Method, req.Amount); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func getInventoryHandler(c *gin.Context) {
	productID := c.Param("product_id")
	level := getInventory(productID)
	c.JSON(http.StatusOK, gin.H{
		"product_id": productID,
		"level":      level,
	})
}

func listOrdersHandler(c *gin.Context) {
	// 模拟数据库查询
	timer := prometheus.NewTimer(dbQueryDuration.WithLabelValues("select", "orders"))
	time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)
	timer.ObserveDuration()

	orders := []Order{
		{ID: "order-1", UserID: "user-1", Amount: 99.99, Category: "electronics", Status: "completed"},
		{ID: "order-2", UserID: "user-2", Amount: 49.99, Category: "books", Status: "pending"},
	}

	c.JSON(http.StatusOK, orders)
}

// ==================== 主程序 ====================

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(PrometheusMiddleware())

	// Prometheus 指标端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 业务 API
	api := r.Group("/api")
	{
		api.GET("/orders", listOrdersHandler)
		api.POST("/orders", createOrderHandler)
		api.POST("/payments", processPaymentHandler)
		api.GET("/inventory/:product_id", getInventoryHandler)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 启动模拟流量生成器
	go generateTraffic()

	fmt.Println("Server starting on :8080")
	fmt.Println("Metrics available at http://localhost:8080/metrics")
	r.Run(":8080")
}

// 生成模拟流量
func generateTraffic() {
	time.Sleep(2 * time.Second)
	client := &http.Client{Timeout: 10 * time.Second}
	categories := []string{"electronics", "books", "clothing", "food"}
	paymentMethods := []string{"credit_card", "paypal", "bank_transfer"}

	for {
		// 模拟创建订单
		go func() {
			category := categories[rand.Intn(len(categories))]
			amount := 10 + rand.Float64()*500
			body := fmt.Sprintf(`{"category":"%s","amount":%.2f}`, category, amount)
			resp, err := client.Post("http://localhost:8080/api/orders", "application/json",
				http.NewRequest("POST", "", nil).Body)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
			_ = body
		}()

		// 模拟支付
		go func() {
			method := paymentMethods[rand.Intn(len(paymentMethods))]
			amount := 10 + rand.Float64()*500
			body := fmt.Sprintf(`{"method":"%s","amount":%.2f}`, method, amount)
			resp, err := client.Post("http://localhost:8080/api/payments", "application/json",
				http.NewRequest("POST", "", nil).Body)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
			_ = body
		}()

		// 模拟查询库存
		go func() {
			productID := fmt.Sprintf("product-%d", rand.Intn(10))
			resp, err := client.Get("http://localhost:8080/api/inventory/" + productID)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
		}()

		time.Sleep(100 * time.Millisecond)
	}
}

/*
运行示例:
  go run main.go

查看指标:
  curl http://localhost:8080/metrics

常用 PromQL 查询:

  # QPS
  rate(myapp_http_requests_total[5m])

  # 按状态分组的 QPS
  sum by (status) (rate(myapp_http_requests_total[5m]))

  # 错误率
  sum(rate(myapp_http_requests_total{status=~"5.."}[5m])) / sum(rate(myapp_http_requests_total[5m]))

  # P95 延迟
  histogram_quantile(0.95, sum by (le) (rate(myapp_http_request_duration_seconds_bucket[5m])))

  # P99 延迟（按路径分组）
  histogram_quantile(0.99, sum by (le, path) (rate(myapp_http_request_duration_seconds_bucket[5m])))

  # 订单金额分布
  histogram_quantile(0.5, rate(myapp_orders_amount_dollars_bucket[5m]))

  # 支付成功率
  sum(rate(myapp_payments_total{status="success"}[5m])) / sum(rate(myapp_payments_total[5m]))

  # 缓存命中率
  sum(rate(myapp_cache_operations_total{result="hit"}[5m])) / sum(rate(myapp_cache_operations_total{operation="get"}[5m]))

  # 数据库查询延迟 P95
  histogram_quantile(0.95, sum by (le, operation) (rate(myapp_db_query_duration_seconds_bucket[5m])))
*/
