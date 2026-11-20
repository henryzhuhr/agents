package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"example.com/gin-layout/internal/entity"
)

const tableNameUser = "t_user"

// userRow 是 t_user 的表映射，只在仓储内部使用。
type userRow struct {
	ID   uint64 `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
}

func (userRow) TableName() string { return tableNameUser }

// UserRepository 描述用户存储能力。
type UserRepository interface {
	// GetByID 返回用户实体副本，未找到时返回 entity.ErrUserNotFound。
	//
	// 返回实体值而不是指针：查询结果要么是实体、要么是错误，调用方无需判空；
	// User 只含值字段，复制成本很低，返回副本也不会暴露仓储内部状态。
	GetByID(ctx context.Context, id uint64) (entity.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 使用已有数据库连接创建用户仓储。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// GetByID implements [UserRepository.GetByID].
func (r *userRepository) GetByID(ctx context.Context, id uint64) (entity.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entity.User{}, entity.ErrUserNotFound
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("query user %d: %w", id, err)
	}
	return entity.User{ID: row.ID, Name: row.Name}, nil
}
