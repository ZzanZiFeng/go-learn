# RabbitMQ 安装配置 (RabbitMQ Setup)

## 概述

RabbitMQ 是一个开源的消息代理，支持 AMQP 协议，提供可靠的消息传递。

## 安装

### macOS

```bash
# 安装
brew install rabbitmq

# 启动服务
brew services start rabbitmq

# 停止服务
brew services stop rabbitmq

# 查看状态
brew services list

# 启用管理插件
rabbitmq-plugins enable rabbitmq_management

# 管理 UI
# http://localhost:15672
# 默认用户: guest / guest
```

### Linux (Ubuntu/Debian)

```bash
# 添加 Erlang 仓库
sudo apt-get install -y erlang

# 添加 RabbitMQ 仓库
curl -s https://packagecloud.io/install/repositories/rabbitmq/rabbitmq-server/script.deb.sh | sudo bash

# 安装
sudo apt-get install -y rabbitmq-server

# 启动
sudo systemctl start rabbitmq-server
sudo systemctl enable rabbitmq-server

# 启用管理插件
sudo rabbitmq-plugins enable rabbitmq_management

# 添加用户
sudo rabbitmqctl add_user admin password123
sudo rabbitmqctl set_user_tags admin administrator
sudo rabbitmqctl set_permissions -p / admin ".*" ".*" ".*"
```

### Docker

```bash
# 基础版
docker run -d --name rabbitmq \
  -p 5672:5672 \
  rabbitmq:3-alpine

# 带管理 UI
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management

# 带持久化
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -v rabbitmq-data:/var/lib/rabbitmq \
  rabbitmq:3-management

# 自定义用户
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=admin \
  -e RABBITMQ_DEFAULT_PASS=admin123 \
  rabbitmq:3-management
```

### Docker Compose

```yaml
version: '3.8'

services:
  rabbitmq:
    image: rabbitmq:3-management
    container_name: rabbitmq
    hostname: rabbitmq
    ports:
      - "5672:5672"    # AMQP
      - "15672:15672"  # Management UI
    environment:
      RABBITMQ_DEFAULT_USER: admin
      RABBITMQ_DEFAULT_PASS: admin123
      RABBITMQ_DEFAULT_VHOST: /
    volumes:
      - rabbitmq-data:/var/lib/rabbitmq
      - ./rabbitmq.conf:/etc/rabbitmq/rabbitmq.conf:ro
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "check_running"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  rabbitmq-data:
```

## 验证安装

```bash
# 检查服务状态
rabbitmqctl status

# 查看节点信息
rabbitmqctl cluster_status

# 列出队列
rabbitmqctl list_queues

# 列出交换机
rabbitmqctl list_exchanges

# 列出用户
rabbitmqctl list_users

# 列出连接
rabbitmqctl list_connections
```

## 管理 UI

访问 `http://localhost:15672`，使用默认用户 `guest/guest` 或自定义用户登录。

### 主要功能

| 功能 | 说明 |
|------|------|
| Overview | 概览、消息速率、节点信息 |
| Connections | 客户端连接列表 |
| Channels | 通道列表 |
| Exchanges | 交换机管理 |
| Queues | 队列管理、消息浏览 |
| Admin | 用户、权限、策略管理 |

### 创建队列

1. 点击 `Queues` 标签
2. 点击 `Add a new queue`
3. 填写队列名称和参数
4. 点击 `Add queue`

### 发送测试消息

1. 点击队列名称进入详情
2. 展开 `Publish message` 部分
3. 填写消息内容
4. 点击 `Publish message`

### 获取消息

1. 在队列详情中展开 `Get messages`
2. 设置获取数量
3. 选择 `Ack mode`
4. 点击 `Get Message(s)`

## 命令行管理

### rabbitmqctl 常用命令

```bash
# 用户管理
rabbitmqctl add_user username password
rabbitmqctl delete_user username
rabbitmqctl change_password username newpass
rabbitmqctl set_user_tags username administrator
rabbitmqctl list_users

# 权限管理
rabbitmqctl set_permissions -p / username ".*" ".*" ".*"
rabbitmqctl clear_permissions -p / username
rabbitmqctl list_permissions

# 虚拟主机
rabbitmqctl add_vhost /dev
rabbitmqctl delete_vhost /dev
rabbitmqctl list_vhosts

# 队列管理
rabbitmqctl list_queues
rabbitmqctl list_queues name messages consumers
rabbitmqctl purge_queue queue_name
rabbitmqctl delete_queue queue_name

# 交换机
rabbitmqctl list_exchanges
rabbitmqctl list_bindings

# 连接/通道
rabbitmqctl list_connections
rabbitmqctl list_channels
rabbitmqctl close_connection <connection_pid>
```

### rabbitmqadmin

```bash
# 安装
wget http://localhost:15672/cli/rabbitmqadmin
chmod +x rabbitmqadmin

# 队列操作
rabbitmqadmin declare queue name=test-queue durable=true
rabbitmqadmin delete queue name=test-queue
rabbitmqadmin list queues

# 交换机操作
rabbitmqadmin declare exchange name=test-exchange type=direct
rabbitmqadmin delete exchange name=test-exchange
rabbitmqadmin list exchanges

# 绑定
rabbitmqadmin declare binding source=test-exchange destination=test-queue routing_key=test

# 发送消息
rabbitmqadmin publish exchange=amq.default routing_key=test-queue payload="Hello"

# 获取消息
rabbitmqadmin get queue=test-queue
```

## 配置文件

### rabbitmq.conf

```ini
# 监听端口
listeners.tcp.default = 5672

# 管理界面
management.tcp.port = 15672

# 默认用户 (生产环境应禁用 guest)
loopback_users.guest = false

# 内存限制 (相对于系统内存)
vm_memory_high_watermark.relative = 0.4

# 磁盘限制
disk_free_limit.absolute = 1GB

# 心跳
heartbeat = 60

# 连接/通道限制
channel_max = 2048
connection_max = infinity

# 日志
log.file.level = info
log.console = true
log.console.level = info

# 集群
# cluster_formation.peer_discovery_backend = rabbit_peer_discovery_classic_config
# cluster_formation.classic_config.nodes.1 = rabbit@node1
# cluster_formation.classic_config.nodes.2 = rabbit@node2
```

### 环境变量

```bash
# RABBITMQ_NODENAME: 节点名称
# RABBITMQ_CONFIG_FILE: 配置文件路径
# RABBITMQ_LOGS: 日志目录
# RABBITMQ_MNESIA_BASE: 数据目录
# RABBITMQ_ENABLED_PLUGINS_FILE: 插件配置文件
```

## 插件管理

```bash
# 列出插件
rabbitmq-plugins list

# 启用插件
rabbitmq-plugins enable rabbitmq_management
rabbitmq-plugins enable rabbitmq_delayed_message_exchange  # 延迟消息
rabbitmq-plugins enable rabbitmq_consistent_hash_exchange  # 一致性哈希
rabbitmq-plugins enable rabbitmq_shovel                    # 数据迁移
rabbitmq-plugins enable rabbitmq_federation                # 联邦

# 禁用插件
rabbitmq-plugins disable rabbitmq_management
```

## 监控

### Prometheus 集成

```bash
# 启用 Prometheus 插件
rabbitmq-plugins enable rabbitmq_prometheus

# 指标端点
# http://localhost:15692/metrics
```

### 关键指标

| 指标 | 说明 |
|------|------|
| queue_messages | 队列消息数 |
| queue_messages_ready | 待消费消息数 |
| queue_messages_unacked | 未确认消息数 |
| channel_consumers | 消费者数 |
| connection_count | 连接数 |
| queue_message_bytes | 消息占用内存 |

### Grafana Dashboard

导入 RabbitMQ 官方 Dashboard ID: `10991`

## 生产环境配置建议

### 内存管理

```ini
# rabbitmq.conf
vm_memory_high_watermark.relative = 0.4     # 内存水位线
vm_memory_high_watermark_paging_ratio = 0.5  # 分页比例
```

### 磁盘管理

```ini
disk_free_limit.absolute = 2GB
```

### 连接管理

```ini
heartbeat = 30
connection_max = 10000
channel_max = 2048
```

### 队列策略

```bash
# 设置最大长度
rabbitmqctl set_policy max-length "^orders\." '{"max-length": 100000}' --apply-to queues

# 设置消息 TTL
rabbitmqctl set_policy message-ttl "^temp\." '{"message-ttl": 86400000}' --apply-to queues

# 设置队列 TTL
rabbitmqctl set_policy queue-ttl "^temp\." '{"expires": 86400000}' --apply-to queues

# 设置死信交换机
rabbitmqctl set_policy dlx "^work\." '{"dead-letter-exchange": "dlx"}' --apply-to queues
```

## 故障排查

### 常见问题

1. **连接失败**
```bash
# 检查服务状态
rabbitmqctl status
# 检查端口
netstat -tlnp | grep 5672
# 检查日志
tail -f /var/log/rabbitmq/rabbit@hostname.log
```

2. **内存告警**
```bash
# 查看内存使用
rabbitmqctl status | grep memory
# 清理不需要的队列
rabbitmqctl delete_queue unused_queue
```

3. **磁盘告警**
```bash
# 查看磁盘使用
df -h
# 清理日志
rabbitmqctl rotate_logs
```

**下一节**：[Go RabbitMQ 客户端](./03-rabbitmq-go.md) - 学习 Go 客户端使用
