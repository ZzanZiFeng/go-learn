# Step 2: API 网关

## 目标

配置 Nginx 作为 API 网关。

## 2.1 Nginx 配置

创建 `gateway/nginx.conf`:

```nginx
upstream auth_service {
    server auth-service:8081;
    keepalive 32;
}

upstream todo_api {
    server todo-api:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name localhost;

    # 静态文件
    location / {
        root /usr/share/nginx/html;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # Auth Service API
    location /api/auth/ {
        proxy_pass http://auth_service/api/auth/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";

        # CORS
        add_header Access-Control-Allow-Origin * always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Authorization, Content-Type" always;

        if ($request_method = OPTIONS) {
            return 204;
        }
    }

    # Auth Service OAuth
    location /api/oauth/ {
        proxy_pass http://auth_service/api/oauth/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Auth Service Users
    location /api/users/ {
        proxy_pass http://auth_service/api/users/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";

        add_header Access-Control-Allow-Origin * always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Authorization, Content-Type" always;

        if ($request_method = OPTIONS) {
            return 204;
        }
    }

    # Todo API
    location /api/todos {
        proxy_pass http://todo_api/api/todos;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";

        add_header Access-Control-Allow-Origin * always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Authorization, Content-Type" always;

        if ($request_method = OPTIONS) {
            return 204;
        }
    }

    # 健康检查
    location /health/auth {
        proxy_pass http://auth_service/health;
        proxy_http_version 1.1;
    }

    location /health/todo {
        proxy_pass http://todo_api/health;
        proxy_http_version 1.1;
    }

    # 错误页面
    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
```

## 2.2 网关功能

### 路由分发

```
请求路径                    转发目标
─────────────────────────────────────────
/api/auth/*         →   Auth Service
/api/oauth/*        →   Auth Service
/api/users/*        →   Auth Service
/api/todos/*        →   Todo API
/api/roles/*        →   Auth Service
/*                  →   静态文件
```

### 负载均衡

```nginx
upstream todo_api {
    least_conn;  # 最少连接算法
    server todo-api-1:8080 weight=3;
    server todo-api-2:8080 weight=2;
    server todo-api-3:8080 weight=1;
    keepalive 32;
}
```

### 限流

```nginx
# 限制请求速率
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

location /api/ {
    limit_req zone=api_limit burst=20 nodelay;
    # ...
}
```

### 缓存

```nginx
# 缓存配置
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=api_cache:10m max_size=100m inactive=60m;

location /api/todos {
    # GET 请求缓存 60 秒
    proxy_cache api_cache;
    proxy_cache_valid 200 60s;
    proxy_cache_methods GET;
    proxy_cache_key "$request_uri|$http_authorization";
    add_header X-Cache-Status $upstream_cache_status;

    # ...
}
```

## 2.3 安全配置

```nginx
# 安全头
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Content-Security-Policy "default-src 'self'" always;

# 隐藏版本
server_tokens off;

# 限制请求体大小
client_max_body_size 10m;

# SSL 配置 (生产环境)
# listen 443 ssl http2;
# ssl_certificate /etc/nginx/ssl/cert.pem;
# ssl_certificate_key /etc/nginx/ssl/key.pem;
# ssl_protocols TLSv1.2 TLSv1.3;
```

## 2.4 日志配置

```nginx
log_format json_combined escape=json '{'
    '"time":"$time_iso8601",'
    '"remote_addr":"$remote_addr",'
    '"method":"$request_method",'
    '"uri":"$request_uri",'
    '"status":$status,'
    '"body_bytes_sent":$body_bytes_sent,'
    '"request_time":$request_time,'
    '"upstream_response_time":"$upstream_response_time",'
    '"user_agent":"$http_user_agent"'
'}';

access_log /var/log/nginx/access.log json_combined;
error_log /var/log/nginx/error.log warn;
```

## 2.5 架构图

```
                         ┌─────────────────┐
                         │    Internet     │
                         └────────┬────────┘
                                  │
                         ┌────────▼────────┐
                         │     Nginx       │
                         │  (API Gateway)  │
                         │    :80/:443     │
                         └────────┬────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
          ▼                       ▼                       ▼
   ┌─────────────┐        ┌─────────────┐        ┌─────────────┐
   │    /api/    │        │  /api/auth  │        │  /api/todos │
   │   静态文件  │        │  /api/users │        │             │
   └─────────────┘        └──────┬──────┘        └──────┬──────┘
          │                      │                      │
          ▼                      ▼                      ▼
   ┌─────────────┐        ┌─────────────┐        ┌─────────────┐
   │   /html     │        │Auth Service │        │  Todo API   │
   │             │        │   :8081     │        │   :8080     │
   └─────────────┘        └─────────────┘        └─────────────┘
```

## 检查点

完成本步骤后：

- [x] 配置 Nginx 反向代理
- [x] 实现路由分发
- [x] 添加 CORS 支持
- [x] 配置安全头

## 下一步

[Step 3: Docker 部署](./step-03-docker.md) - 容器化部署。
