package user

import (
	"context"
	"errors"

	"example.com/gin-layout/internal/entity"
	"example.com/gin-layout/internal/svc"
)

var ErrInvalidID = errors.New("invalid user ID")

type ProfileService struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileService(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileService {
	return &ProfileService{ctx: ctx, svcCtx: svcCtx}
}

func (s *ProfileService) Get(id uint64) (entity.User, error) {
	if id == 0 {
		return entity.User{}, ErrInvalidID
	}
	return s.svcCtx.Repo.User.GetByID(s.ctx, id)
}
