package model

import "errors"

// ErrUserNotFound 表示查询的用户不存在，调用方可用 errors.Is 判断。
var ErrUserNotFound = errors.New("user not found")

// User 是业务实体，不携带 HTTP 或数据库映射标签。
//
// 字段均为值类型，按值返回或传递不会共享可变状态，仓储可直接返回副本。
type User struct {
	ID   uint64
	Name string
}
