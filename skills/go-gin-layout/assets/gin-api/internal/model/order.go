package model

type Order struct {
	ID     uint64
	UserID uint64
	Items  []OrderItem
}
