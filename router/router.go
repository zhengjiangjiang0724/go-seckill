package router

import (
	"go-seckill/controller"
	"go-seckill/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(seckillController *controller.SeckillController) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.RateLimitMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		// 商品相关
		api.GET("/products", seckillController.GetProducts)
		api.GET("/products/:id", seckillController.GetProduct)

		// 秒杀相关（需要用户级限流）
		seckill := api.Group("/seckill")
		seckill.Use(middleware.UserRateLimitMiddleware())
		{
			seckill.POST("/token", seckillController.GenerateToken)
			seckill.POST("/buy", seckillController.Seckill)
		}

		// 订单相关
		api.GET("/orders/:orderNo", seckillController.GetOrder)

		// 管理接口
		admin := api.Group("/admin")
		{
			admin.POST("/products", seckillController.CreateProduct)
			admin.PUT("/orders/status", seckillController.UpdateOrderStatus)
		}
	}

	return r
}

