package model

// Order 是业务实体。按值传递只复制 Items 的切片头，
// 需要与调用方隔离底层数组时由仓储显式复制。
type Order struct {
	ID     uint64
	UserID uint64
	Items  []OrderItem
}
