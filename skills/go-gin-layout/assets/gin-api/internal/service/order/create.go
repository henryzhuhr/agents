package order

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"example.com/gin-layout/internal/model"
	"example.com/gin-layout/internal/svc"
)

const (
	maxItems    = 100
	maxSKU      = 64
	maxQuantity = math.MaxInt32
)

var ErrInvalidOrder = errors.New("invalid order")

type CreateService struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateService(ctx context.Context, svcCtx *svc.ServiceContext) *CreateService {
	return &CreateService{ctx: ctx, svcCtx: svcCtx}
}

func (s *CreateService) Create(userID uint64, items []model.OrderItem) (model.Order, error) {
	if userID == 0 || len(items) == 0 || len(items) > maxItems {
		return model.Order{}, ErrInvalidOrder
	}
	// 使用独立切片，避免规范化输入时修改调用方的数据。
	normalized := make([]model.OrderItem, len(items))
	for i, item := range items {
		item.SKU = strings.TrimSpace(item.SKU)
		if item.SKU == "" || len([]rune(item.SKU)) > maxSKU || item.Quantity <= 0 || item.Quantity > maxQuantity {
			return model.Order{}, ErrInvalidOrder
		}
		normalized[i] = item
	}
	if _, err := s.svcCtx.Repo.User.GetByID(s.ctx, userID); err != nil {
		return model.Order{}, fmt.Errorf("read ordering user: %w", err)
	}
	return s.svcCtx.Repo.Order.Create(s.ctx, model.Order{UserID: userID, Items: normalized})
}
