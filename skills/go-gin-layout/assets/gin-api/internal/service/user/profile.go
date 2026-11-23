package user

import (
	"context"
	"errors"

	"example.com/gin-layout/internal/model"
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

func (s *ProfileService) Get(id uint64) (model.User, error) {
	if id == 0 {
		return model.User{}, ErrInvalidID
	}
	return s.svcCtx.Repo.User.GetByID(s.ctx, id)
}
