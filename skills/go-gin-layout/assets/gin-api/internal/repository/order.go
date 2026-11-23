package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"example.com/gin-layout/internal/model"
)

const (
	tableNameOrder     = "t_order"
	tableNameOrderItem = "t_order_item"
)

type orderRow struct {
	ID     uint64 `gorm:"column:id;primaryKey"`
	UserID uint64 `gorm:"column:user_id"`
}

func (orderRow) TableName() string { return tableNameOrder }

type orderItemRow struct {
	ID       uint64 `gorm:"column:id;primaryKey"`
	OrderID  uint64 `gorm:"column:order_id"`
	SKU      string `gorm:"column:sku"`
	Quantity int    `gorm:"column:quantity"`
}

func (orderItemRow) TableName() string { return tableNameOrderItem }

// OrderRepository 描述订单存储能力。
type OrderRepository interface {
	// Create 保存订单并返回持久化后的副本；订单与明细在同一事务中写入，
	// 返回值的 Items 与仓储内部切片不共享底层数组。
	Create(ctx context.Context, order model.Order) (model.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 使用已有数据库连接创建订单仓储。
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Create implements [OrderRepository.Create].
func (r *orderRepository) Create(ctx context.Context, order model.Order) (model.Order, error) {
	if err := ctx.Err(); err != nil {
		return model.Order{}, err
	}
	var created model.Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := orderRow{UserID: order.UserID}
		if err := tx.Create(&row).Error; err != nil {
			return fmt.Errorf("insert order: %w", err)
		}
		items := make([]orderItemRow, len(order.Items))
		for i, item := range order.Items {
			items[i] = orderItemRow{OrderID: row.ID, SKU: item.SKU, Quantity: item.Quantity}
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("insert order items: %w", err)
			}
		}
		created = model.Order{
			ID:     row.ID,
			UserID: row.UserID,
			Items:  itemEntities(items),
		}
		return nil
	})
	if err != nil {
		return model.Order{}, err
	}
	return created, nil
}

func itemEntities(rows []orderItemRow) []model.OrderItem {
	items := make([]model.OrderItem, len(rows))
	for i, row := range rows {
		items[i] = model.OrderItem{SKU: row.SKU, Quantity: row.Quantity}
	}
	return items
}
