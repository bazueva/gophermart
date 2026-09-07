package entities

import "time"

type OrderStatus string

const (
	OrdersStatusNew        OrderStatus = "NEW"
	OrdersStatusProcessing OrderStatus = "PROCESSING"
	OrdersStatusInvalid    OrderStatus = "INVALID"
	OrdersStatusProcessed  OrderStatus = "PROCESSED"
)

// Order структура заказ.
type Order struct {
	ID          int32
	OrderID     string
	Status      OrderStatus
	UserID      int32
	BonusSum    float64
	CreatedAt   *time.Time
	ProcessedAt *time.Time
	NextCheckAt *time.Time
}

// BalanceWithdraw структура снятия бонусов.
type BalanceWithdraw struct {
	Order string
	Sum   float64
}

// OrderFilterOrderType тип фильтрации заказов.
type OrderFilterOrderType int

const (
	// OrderFilterAddBalanceType Тип заказа добавления бонусов.
	OrderFilterAddBalanceType OrderFilterOrderType = iota
	// OrderFilterWriteOffBalanceType Тип заказа снятия бонусов.
	OrderFilterWriteOffBalanceType
)

// OrderFilter структура фильтрации заказов.
type OrderFilter struct {
	OrderType *OrderFilterOrderType
	UserID    int32
	Statuses  []OrderStatus
}
