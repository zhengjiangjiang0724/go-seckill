package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go-seckill/cache"
	"go-seckill/config"
	"go-seckill/database"
	"go-seckill/models"
	"go-seckill/utils"
)

type SeckillService struct {
	cfg *config.Config
}

func NewSeckillService(cfg *config.Config) *SeckillService {
	return &SeckillService{cfg: cfg}
}

// PreheatStock 预热库存到Redis
func (s *SeckillService) PreheatStock(productID uint, stock int) error {
	key := fmt.Sprintf("%s%d", s.cfg.Seckill.StockPrefix, productID)
	return cache.Set(key, stock, time.Duration(s.cfg.Seckill.TokenExpire)*time.Second)
}

// GetStockFromRedis 从Redis获取库存
func (s *SeckillService) GetStockFromRedis(productID uint) (int64, error) {
	key := fmt.Sprintf("%s%d", s.cfg.Seckill.StockPrefix, productID)
	stockStr, err := cache.Get(key)
	if err != nil {
		return 0, err
	}
	var stock int64
	fmt.Sscanf(stockStr, "%d", &stock)
	return stock, nil
}

// GenerateToken 生成秒杀令牌
func (s *SeckillService) GenerateToken(userID string, productID uint) (string, error) {
	// 检查是否在秒杀时间
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return "", errors.New("product not found")
	}

	if !utils.IsSeckillTime(product.StartTime, product.EndTime) {
		return "", errors.New("seckill not started or ended")
	}

	// 检查库存
	stock, err := s.GetStockFromRedis(productID)
	if err != nil || stock <= 0 {
		return "", errors.New("out of stock")
	}

	// 生成令牌
	token := fmt.Sprintf("%s-%d-%d", userID, productID, time.Now().UnixNano())
	tokenKey := fmt.Sprintf("%s%s", s.cfg.Seckill.TokenPrefix, token)
	
	// 令牌有效期1小时
	if err := cache.Set(tokenKey, "1", time.Duration(s.cfg.Seckill.TokenExpire)*time.Second); err != nil {
		return "", err
	}

	return token, nil
}

// Seckill 秒杀核心逻辑（使用Lua脚本保证原子性）
func (s *SeckillService) Seckill(userID string, productID uint, token string) (*models.Order, error) {
	// 验证令牌
	tokenKey := fmt.Sprintf("%s%s", s.cfg.Seckill.TokenPrefix, token)
	_, err := cache.Get(tokenKey)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// 使用Lua脚本保证原子性：检查库存 -> 扣减库存 -> 生成订单号
	luaScript := `
		local stockKey = KEYS[1]
		local orderKey = KEYS[2]
		local stock = tonumber(redis.call('get', stockKey) or 0)
		
		if stock <= 0 then
			return {0, 'out of stock'}
		end
		
		redis.call('decr', stockKey)
		local orderNo = ARGV[1]
		redis.call('setex', orderKey, 3600, orderNo)
		
		return {1, orderNo}
	`

	stockKey := fmt.Sprintf("%s%d", s.cfg.Seckill.StockPrefix, productID)
	orderKey := fmt.Sprintf("%s%s:%d", s.cfg.Seckill.OrderPrefix, userID, productID)
	orderNo := utils.GenerateOrderNo()

	result, err := cache.Eval(luaScript, []string{stockKey, orderKey}, orderNo)
	if err != nil {
		return nil, fmt.Errorf("seckill failed: %w", err)
	}

	// 处理Lua脚本返回值
	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) < 2 {
		return nil, errors.New("invalid lua script result")
	}

	success, ok := resultArray[0].(int64)
	if !ok {
		return nil, errors.New("invalid lua script result type")
	}

	if success == 0 {
		errMsg, _ := resultArray[1].(string)
		if errMsg == "" {
			errMsg = "seckill failed"
		}
		return nil, errors.New(errMsg)
	}

	// 获取商品信息
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// 创建订单
	order := &models.Order{
		OrderNo:     orderNo,
		UserID:      userID,
		ProductID:   productID,
		ProductName: product.Name,
		Price:       product.Price,
		Status:      models.OrderStatusPending,
	}

	// 使用分布式锁保护数据库写入
	lockKey := fmt.Sprintf("%sorder:%s", s.cfg.Seckill.LockPrefix, orderNo)
	lock := utils.NewDistributedLock(lockKey, 5*time.Second)
	
	locked, err := lock.TryLockWithRetry(3, 100*time.Millisecond)
	if !locked {
		return nil, errors.New("failed to acquire lock")
	}
	defer lock.Unlock()

	if err := database.DB.Create(order).Error; err != nil {
		log.Printf("Failed to create order: %v", err)
		// 回滚库存
		cache.Incr(stockKey)
		return nil, errors.New("failed to create order")
	}

	// 删除令牌
	cache.Del(tokenKey)

	return order, nil
}

// CheckUserOrder 检查用户是否已经下过单
func (s *SeckillService) CheckUserOrder(userID string, productID uint) (bool, error) {
	orderKey := fmt.Sprintf("%s%s:%d", s.cfg.Seckill.OrderPrefix, userID, productID)
	_, err := cache.Get(orderKey)
	if err == nil {
		return true, nil
	}

	// 检查数据库
	var count int64
	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND product_id = ? AND status != ?", userID, productID, models.OrderStatusCancelled).
		Count(&count)
	
	return count > 0, nil
}

// GetProduct 获取商品信息
func (s *SeckillService) GetProduct(productID uint) (*models.Product, error) {
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// ListProducts 获取商品列表
func (s *SeckillService) ListProducts() ([]models.Product, error) {
	var products []models.Product
	if err := database.DB.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// CreateProduct 创建商品
func (s *SeckillService) CreateProduct(product *models.Product) error {
	if err := database.DB.Create(product).Error; err != nil {
		return err
	}
	// 预热库存到Redis
	return s.PreheatStock(product.ID, product.SeckillStock)
}

// GetOrder 获取订单信息
func (s *SeckillService) GetOrder(orderNo string) (*models.Order, error) {
	var order models.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrderStatus 更新订单状态
func (s *SeckillService) UpdateOrderStatus(orderNo string, status string) error {
	return database.DB.Model(&models.Order{}).
		Where("order_no = ?", orderNo).
		Update("status", status).Error
}

