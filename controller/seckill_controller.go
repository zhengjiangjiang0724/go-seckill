package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-seckill/models"
	"go-seckill/service"
)

type SeckillController struct {
	seckillService *service.SeckillService
}

func NewSeckillController(seckillService *service.SeckillService) *SeckillController {
	return &SeckillController{
		seckillService: seckillService,
	}
}

// Response 统一响应格式
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// GetProducts 获取商品列表
func (c *SeckillController) GetProducts(ctx *gin.Context) {
	products, err := c.seckillService.ListProducts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: products,
	})
}

// GetProduct 获取商品详情
func (c *SeckillController) GetProduct(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "invalid product id",
		})
		return
	}

	product, err := c.seckillService.GetProduct(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, Response{
			Code: 404,
			Msg:  "product not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: product,
	})
}

// GenerateToken 生成秒杀令牌
func (c *SeckillController) GenerateToken(ctx *gin.Context) {
	var req struct {
		ProductID uint   `json:"product_id" binding:"required"`
		UserID    string `json:"user_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  err.Error(),
		})
		return
	}

	token, err := c.seckillService.GenerateToken(req.UserID, req.ProductID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: gin.H{"token": token},
	})
}

// Seckill 秒杀接口
func (c *SeckillController) Seckill(ctx *gin.Context) {
	var req struct {
		ProductID uint   `json:"product_id" binding:"required"`
		UserID    string `json:"user_id" binding:"required"`
		Token     string `json:"token" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  err.Error(),
		})
		return
	}

	// 检查用户是否已经下过单
	hasOrder, err := c.seckillService.CheckUserOrder(req.UserID, req.ProductID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  err.Error(),
		})
		return
	}

	if hasOrder {
		ctx.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "user already has an order",
		})
		return
	}

	order, err := c.seckillService.Seckill(req.UserID, req.ProductID, req.Token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "seckill success",
		Data: order,
	})
}

// GetOrder 获取订单信息
func (c *SeckillController) GetOrder(ctx *gin.Context) {
	orderNo := ctx.Param("orderNo")
	
	order, err := c.seckillService.GetOrder(orderNo)
	if err != nil {
		ctx.JSON(http.StatusNotFound, Response{
			Code: 404,
			Msg:  "order not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: order,
	})
}

// CreateProduct 创建商品（管理接口）
func (c *SeckillController) CreateProduct(ctx *gin.Context) {
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  err.Error(),
		})
		return
	}

	if err := c.seckillService.CreateProduct(&product); err != nil {
		ctx.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "product created successfully",
		Data: product,
	})
}

// UpdateOrderStatus 更新订单状态（管理接口）
func (c *SeckillController) UpdateOrderStatus(ctx *gin.Context) {
	var req struct {
		OrderNo string `json:"order_no" binding:"required"`
		Status  string `json:"status" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  err.Error(),
		})
		return
	}

	if err := c.seckillService.UpdateOrderStatus(req.OrderNo, req.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "order status updated successfully",
	})
}

