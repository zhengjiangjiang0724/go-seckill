package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"go-seckill/cache"
	"go-seckill/config"
	"go-seckill/controller"
	"go-seckill/database"
	"go-seckill/router"
	"go-seckill/service"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化Redis
	if err := cache.InitRedis(cfg); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化服务
	seckillService := service.NewSeckillService(cfg)

	// 初始化控制器
	seckillController := controller.NewSeckillController(seckillService)

	// 设置路由
	r := router.SetupRouter(seckillController)

	// 启动服务
	addr := ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

