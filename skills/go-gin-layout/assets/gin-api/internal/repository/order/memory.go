package order

import (
	"context"
	"sync"

	"example.com/gin-layout/internal/model"
)

type Repository struct {
	mu     sync.Mutex
	nextID uint64
	orders map[uint64]model.Order
}

func New() *Repository {
	return &Repository{nextID: 1, orders: make(map[uint64]model.Order)}
}

func (r *Repository) Create(ctx context.Context, order model.Order) (model.Order, error) {
	if err := ctx.Err(); err != nil {
		return model.Order{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return model.Order{}, err
	}
	order.ID = r.nextID
	r.nextID++
	// 内存仓库与返回值各持有独立切片，避免调用方修改已存储数据。
	order.Items = append([]model.OrderItem(nil), order.Items...)
	r.orders[order.ID] = order
	result := order
	result.Items = append([]model.OrderItem(nil), order.Items...)
	return result, nil
}
