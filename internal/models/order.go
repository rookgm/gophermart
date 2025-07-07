package models

import "time"

//NEW — заказ загружен в систему, но не попал в обработку;
//PROCESSING — вознаграждение за заказ рассчитывается;
//INVALID — система расчёта вознаграждений отказала в расчёте;
//PROCESSED — данные по заказу проверены и информация о расчёте успешно получена.

// order status
const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

// Order is order entity
type Order struct {
	ID        uint64
	UserID    uint64
	OrderID   string
	Status    string
	Accrual   *float64
	CreatedAt time.Time
}
