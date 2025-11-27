# Go高并发秒杀系统

一个基于Go语言开发的高并发秒杀系统，支持高并发场景下的商品秒杀功能。

## 项目特性

- 🚀 **高性能**: 支持10,000+ QPS并发请求
- 🔒 **数据一致性**: 使用Lua脚本和分布式锁，确保不超卖
- 🛡️ **系统稳定性**: 多层限流机制，保护系统不被过载
- 🎫 **令牌机制**: 提前过滤无效请求，优化用户体验
- 📦 **库存预热**: Redis缓存库存，减少数据库压力
- 🔐 **防刷机制**: 用户级限流，防止刷单

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin 1.9.1
- **数据库**: MySQL 8.0+
- **缓存**: Redis 8.0+
- **ORM**: GORM 1.25+

## 项目结构

```
go-seckill/
├── cache/              # Redis缓存封装
├── config/             # 配置管理
├── controller/         # 控制器层
├── database/           # 数据库连接
├── docs/               # 文档
│   ├── architecture.md      # 架构设计文档
│   └── technical_summary.md # 技术总结文档
├── middleware/         # 中间件（限流等）
├── models/             # 数据模型
├── router/             # 路由配置
├── service/            # 业务逻辑层
├── tests/              # 测试代码
│   ├── benchmark_test.go    # 性能测试
│   ├── load_test.sh         # 压力测试脚本
│   ├── wrk_test.lua         # wrk测试脚本
│   └── performance_report.md # 性能测试报告
├── utils/              # 工具函数
├── go.mod              # 依赖管理
├── main.go             # 入口文件
└── README.md           # 项目说明
```

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 8.0+

### 安装依赖

```bash
go mod download
```

### 配置环境变量

```bash
export SERVER_PORT=8080
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=seckill
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=
```

### 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE seckill CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

数据库表会在应用启动时自动创建（通过GORM AutoMigrate）。

### 运行项目

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## API接口

### 商品相关

#### 获取商品列表
```http
GET /api/v1/products
```

#### 获取商品详情
```http
GET /api/v1/products/:id
```

#### 创建商品（管理接口）
```http
POST /api/v1/admin/products
Content-Type: application/json

{
  "name": "秒杀商品",
  "price": 99.99,
  "stock": 10000,
  "seckill_stock": 1000,
  "start_time": "2024-01-01T10:00:00Z",
  "end_time": "2024-01-01T12:00:00Z"
}
```

### 秒杀相关

#### 生成秒杀令牌
```http
POST /api/v1/seckill/token
Content-Type: application/json

{
  "user_id": "user123",
  "product_id": 1
}
```

#### 执行秒杀
```http
POST /api/v1/seckill/buy
Content-Type: application/json

{
  "user_id": "user123",
  "product_id": 1,
  "token": "user123-1-1234567890"
}
```

### 订单相关

#### 查询订单
```http
GET /api/v1/orders/:orderNo
```

## 核心实现

### 1. 秒杀令牌机制

用户在秒杀开始前先获取令牌，令牌存储在Redis中，设置TTL。执行秒杀时验证令牌有效性，提前过滤无效请求。

### 2. 库存扣减 - Lua脚本

使用Redis Lua脚本将检查库存和扣减库存打包成一个原子操作，确保不超卖：

```lua
local stockKey = KEYS[1]
local stock = tonumber(redis.call('get', stockKey) or 0)

if stock <= 0 then
    return {0, 'out of stock'}
end

redis.call('decr', stockKey)
return {1, 'success'}
```

### 3. 分布式锁

使用Redis的SETNX命令实现分布式锁，保护数据库订单创建的临界区：

```go
lock := utils.NewDistributedLock(lockKey, 5*time.Second)
locked, _ := lock.TryLockWithRetry(3, 100*time.Millisecond)
defer lock.Unlock()
```

### 4. 限流机制

实现两层限流：
- **全局限流**: 令牌桶算法，容量10000，速率1000/秒
- **用户限流**: 每个用户独立的令牌桶，限制5次/秒

## 性能测试

### 运行性能测试

```bash
cd tests
go test -v -run TestConcurrentSeckill
```

### 使用wrk进行压力测试

```bash
wrk -t4 -c100 -d30s -s tests/wrk_test.lua http://localhost:8080
```

详细测试报告请参考 [tests/performance_report.md](tests/performance_report.md)

## 文档

- [架构设计文档](docs/architecture.md) - 详细的系统架构设计
- [技术总结文档](docs/technical_summary.md) - 技术实现细节和挑战解决方案
- [性能测试报告](tests/performance_report.md) - 性能测试结果和分析

## 项目亮点

### 1. 高并发处理
- Redis缓存库存，减少数据库压力
- Lua脚本保证原子操作
- 令牌机制提前过滤无效请求

### 2. 数据一致性
- Lua脚本确保库存扣减原子性
- 分布式锁保护数据库写入
- 防止超卖和重复下单

### 3. 系统稳定性
- 多层限流保护系统
- 降级策略保证可用性
- 完善的错误处理

## 遇到的问题和解决方案

### 问题1: 超卖问题
**解决方案**: 使用Redis Lua脚本将检查库存和扣减库存打包成原子操作

### 问题2: 数据库压力过大
**解决方案**: Redis缓存库存和令牌，减少数据库查询

### 问题3: 系统过载
**解决方案**: 多层限流机制（全局限流 + 用户限流）

详细的问题和解决方案请参考 [技术总结文档](docs/technical_summary.md)

## 开发计划

- [ ] 消息队列异步处理订单
- [ ] 支付系统集成
- [ ] 订单查询优化
- [ ] 数据分析统计
- [ ] 微服务拆分

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交Issue。

