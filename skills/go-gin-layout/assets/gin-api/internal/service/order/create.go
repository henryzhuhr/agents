package order

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"example.com/gin-layout/internal/entity"
	"example.com/gin-layout/internal/svc"
)

var ErrInvalidOrder = errors.New("invalid order")

type CreateService struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateService(ctx context.Context, svcCtx *svc.ServiceContext) *CreateService {
	return &CreateService{ctx: ctx, svcCtx: svcCtx}
}

func (s *CreateService) Create(userID uint64, items []entity.OrderItem) (entity.Order, error) {
	if userID == 0 || len(items) == 0 {
		return entity.Order{}, ErrInvalidOrder
	}
	// 使用独立切片，避免规范化输入时修改调用方的数据。
	normalized := make([]entity.OrderItem, len(items))
	for i, item := range items {
		item.SKU = strings.TrimSpace(item.SKU)
		if item.SKU == "" || item.Quantity <= 0 {
			return entity.Order{}, ErrInvalidOrder
		}
		normalized[i] = item
	}
	if _, err := s.svcCtx.Repo.User.GetByID(s.ctx, userID); err != nil {
		return entity.Order{}, fmt.Errorf("read ordering user: %w", err)
	}
	return s.svcCtx.Repo.Order.Create(s.ctx, entity.Order{UserID: userID, Items: normalized})
}
