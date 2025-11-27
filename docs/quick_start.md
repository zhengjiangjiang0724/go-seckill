# 快速启动指南

## 前置条件

- Go 1.21+
- MySQL 8.0+
- Redis 8.0+
- Make（可选）

## 方式一：直接运行

### 1. 安装依赖

```bash
cd go-seckill
go mod download
```

### 2. 配置环境变量

复制环境变量示例文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置数据库和Redis连接信息。

或者直接设置环境变量：

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=seckill
export REDIS_ADDR=localhost:6379
```

### 3. 初始化数据库

```bash
mysql -u root -p < scripts/init_db.sql
```

或者让应用自动创建（GORM会自动迁移表结构）。

### 4. 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## 方式二：使用 Docker Compose

### 1. 启动所有服务

```bash
docker-compose up -d
```

这将启动：
- MySQL (端口 3306)
- Redis (端口 6379)
- 秒杀应用 (端口 8080)

### 2. 查看日志

```bash
docker-compose logs -f seckill
```

### 3. 停止服务

```bash
docker-compose down
```

## 方式三：使用 Makefile

### 1. 下载依赖

```bash
make deps
```

### 2. 构建项目

```bash
make build
```

### 3. 运行项目

```bash
make run
```

### 4. 运行测试

```bash
make test
```

### 5. 运行性能测试

```bash
make bench
```

## 测试API

### 1. 创建商品

```bash
curl -X POST http://localhost:8080/api/v1/admin/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试秒杀商品",
    "price": 99.99,
    "stock": 10000,
    "seckill_stock": 1000,
    "start_time": "2024-01-01T10:00:00Z",
    "end_time": "2024-01-01T12:00:00Z"
  }'
```

### 2. 获取商品列表

```bash
curl http://localhost:8080/api/v1/products
```

### 3. 生成秒杀令牌

```bash
curl -X POST http://localhost:8080/api/v1/seckill/token \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "product_id": 1
  }'
```

### 4. 执行秒杀

```bash
curl -X POST http://localhost:8080/api/v1/seckill/buy \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "product_id": 1,
    "token": "your_token_here"
  }'
```

## 性能测试

### 使用Go测试框架

```bash
cd tests
go test -v -run TestConcurrentSeckill
```

### 使用wrk

```bash
# 安装wrk
# macOS: brew install wrk
# Linux: apt-get install wrk

# 运行测试
wrk -t4 -c100 -d30s -s tests/wrk_test.lua http://localhost:8080
```

### 使用Apache Bench

```bash
# 测试商品列表接口
ab -n 10000 -c 100 http://localhost:8080/api/v1/products
```

## 健康检查

```bash
curl http://localhost:8080/health
```

应该返回：

```json
{"status":"ok"}
```

## 常见问题

### 1. 数据库连接失败

检查：
- MySQL服务是否启动
- 数据库用户和密码是否正确
- 数据库是否存在

### 2. Redis连接失败

检查：
- Redis服务是否启动
- Redis地址和端口是否正确
- Redis是否需要密码

### 3. 端口被占用

修改环境变量 `SERVER_PORT` 或修改 `docker-compose.yml` 中的端口映射。

## 下一步

- 阅读 [架构设计文档](architecture.md)
- 阅读 [技术总结文档](technical_summary.md)
- 查看 [性能测试报告](../tests/performance_report.md)

