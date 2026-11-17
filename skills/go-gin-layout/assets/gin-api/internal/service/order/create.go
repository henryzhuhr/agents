package order

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"example.com/gin-layout/internal/model"
)

var ErrInvalidOrder = errors.New("invalid order")

type Repository interface {
	Create(context.Context, model.Order) (model.Order, error)
}

// UserReader 只描述订单用例需要的能力，避免依赖用户 Service 的实现。
type UserReader interface {
	Get(context.Context, uint64) (model.User, error)
}

type Service struct {
	repo  Repository
	users UserReader
}

func New(repo Repository, users UserReader) *Service {
	return &Service{repo: repo, users: users}
}

func (s *Service) Create(ctx context.Context, userID uint64, items []model.OrderItem) (model.Order, error) {
	if userID == 0 || len(items) == 0 {
		return model.Order{}, ErrInvalidOrder
	}
	// 使用独立切片，避免规范化输入时修改调用方的数据。
	normalized := make([]model.OrderItem, len(items))
	for i, item := range items {
		item.SKU = strings.TrimSpace(item.SKU)
		if item.SKU == "" || item.Quantity <= 0 {
			return model.Order{}, ErrInvalidOrder
		}
		normalized[i] = item
	}
	if _, err := s.users.Get(ctx, userID); err != nil {
		return model.Order{}, fmt.Errorf("read ordering user: %w", err)
	}
	return s.repo.Create(ctx, model.Order{UserID: userID, Items: normalized})
}
