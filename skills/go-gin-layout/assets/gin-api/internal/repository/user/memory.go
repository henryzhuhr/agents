package user

import (
	"context"

	"example.com/gin-layout/internal/model"
)

// Repository 仅维护不可变的示例用户；替换存储不会影响 Service 接口。
type Repository struct {
	users map[uint64]model.User
}

func New() *Repository {
	return &Repository{users: map[uint64]model.User{
		1: {ID: 1, Name: "Alice"},
	}}
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (model.User, error) {
	if err := ctx.Err(); err != nil {
		return model.User{}, err
	}
	user, ok := r.users[id]
	if !ok {
		return model.User{}, model.ErrUserNotFound
	}
	return user, nil
}
