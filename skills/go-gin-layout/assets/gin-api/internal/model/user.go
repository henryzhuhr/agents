package model

import "errors"

var ErrUserNotFound = errors.New("user not found")

// User 是业务实体，不携带 HTTP 或数据库映射标签。
type User struct {
	ID   uint64
	Name string
}
