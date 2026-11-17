package user

import (
	"context"
	"errors"

	"example.com/gin-layout/internal/model"
)

var ErrInvalidID = errors.New("invalid user ID")

// Repository 由消费方定义，Service 无需导入具体存储实现。
type Repository interface {
	GetByID(context.Context, uint64) (model.User, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, id uint64) (model.User, error) {
	if id == 0 {
		return model.User{}, ErrInvalidID
	}
	return s.repo.GetByID(ctx, id)
}
