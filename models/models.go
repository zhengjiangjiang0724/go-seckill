package models

import (
	"time"

	"gorm.io/gorm"
)

// Product 商品模型
type Product struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock       int       `gorm:"type:int;not null;default:0" json:"stock"`
	StartTime   time.Time `gorm:"type:datetime;not null" json:"start_time"`
	EndTime     time.Time `gorm:"type:datetime;not null" json:"end_time"`
	SeckillStock int     `gorm:"type:int;not null;default:0" json:"seckill_stock"`
}

// Order 订单模型
type Order struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	OrderNo     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"`
	UserID      string    `gorm:"type:varchar(64);not null;index" json:"user_id"`
	ProductID   uint      `gorm:"type:int;not null;index" json:"product_id"`
	ProductName string    `gorm:"type:varchar(255);not null" json:"product_name"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Status      string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Product     Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// OrderStatus 订单状态常量
const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusCancelled = "cancelled"
	OrderStatusCompleted = "completed"
)

