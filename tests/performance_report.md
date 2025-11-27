# 性能测试报告

## 测试环境

- **服务器配置**:
  - CPU: 4核
  - 内存: 8GB
  - 操作系统: Linux/macOS
  - Go版本: 1.21+

- **依赖服务**:
  - MySQL 8.0+
  - Redis: 8.0+

## 测试场景

### 场景1: 获取商品列表

**测试参数**:
- 并发数: 100
- 总请求数: 10,000
- 接口: `GET /api/v1/products`

**预期结果**:
- QPS: > 5,000
- 平均响应时间: < 50ms
- 99%响应时间: < 100ms

### 场景2: 生成秒杀令牌

**测试参数**:
- 并发数: 500
- 总请求数: 50,000
- 接口: `POST /api/v1/seckill/token`

**预期结果**:
- QPS: > 3,000
- 平均响应时间: < 100ms
- 成功率: > 95%

### 场景3: 并发秒杀

**测试参数**:
- 并发数: 1,000
- 总请求数: 10,000
- 库存: 1,000
- 接口: `POST /api/v1/seckill/buy`

**预期结果**:
- 成功订单数: 1,000（不超过库存）
- 平均响应时间: < 200ms
- 库存一致性: 100%（无超卖）

## 测试执行

### 使用 Go 测试框架

```bash
# 运行并发测试
cd tests
go test -v -run TestConcurrentSeckill

# 运行基准测试
go test -v -bench=BenchmarkSeckill -benchmem
```

### 使用 Apache Bench

```bash
# 获取商品列表
ab -n 10000 -c 100 http://localhost:8080/api/v1/products

# 生成令牌（需要POST请求，使用curl脚本）
bash tests/load_test.sh
```

### 使用 wrk

```bash
# 安装wrk
# macOS: brew install wrk
# Linux: yum install wrk 或 apt-get install wrk

# 运行测试
wrk -t4 -c100 -d30s -s tests/wrk_test.lua http://localhost:8080
```

## 性能指标

### 1. 吞吐量 (QPS)

| 接口 | 目标QPS | 实际QPS | 说明 |
|------|---------|---------|------|
| GET /api/v1/products | 5,000 | - | 需要实际测试 |
| POST /api/v1/seckill/token | 3,000 | - | 需要实际测试 |
| POST /api/v1/seckill/buy | 1,000 | - | 需要实际测试 |

### 2. 响应时间

| 接口 | 平均响应时间 | P95响应时间 | P99响应时间 |
|------|-------------|-------------|-------------|
| GET /api/v1/products | < 50ms | < 100ms | < 200ms |
| POST /api/v1/seckill/token | < 100ms | < 200ms | < 500ms |
| POST /api/v1/seckill/buy | < 200ms | < 500ms | < 1000ms |

### 3. 系统资源

| 资源 | 使用率 | 说明 |
|------|--------|------|
| CPU | < 70% | 峰值不超过80% |
| 内存 | < 2GB | 包括Go运行时 |
| Redis连接数 | < 100 | 连接池大小 |
| MySQL连接数 | < 50 | 连接池大小 |

## 优化建议

### 1. 缓存优化

- 商品信息缓存：减少数据库查询
- 使用Redis Pipeline减少网络往返
- 预热热点数据到Redis

### 2. 数据库优化

- 使用读写分离
- 订单表分表（按时间或用户ID）
- 添加适当的索引

### 3. 代码优化

- 使用连接池
- 减少不必要的序列化/反序列化
- 使用协程池控制并发

### 4. 架构优化

- 使用消息队列异步处理订单
- 实施服务降级和熔断
- 使用CDN加速静态资源

## 测试结果示例

```
Concurrent Seckill Test Results:
  Total Requests: 1000
  Success: 1000
  Failed: 0
  Duration: 2.345s
  QPS: 426.44
  Average Latency: 2.345ms
```

## 注意事项

1. **测试前准备**:
   - 确保MySQL和Redis服务正常运行
   - 创建测试商品并预热库存
   - 清理之前的测试数据

2. **测试数据**:
   - 使用不同的用户ID避免限流
   - 确保商品库存充足
   - 测试结束后清理数据

3. **监控指标**:
   - 实时监控CPU、内存使用情况
   - 监控Redis和MySQL的连接数和QPS
   - 记录错误日志和慢查询

