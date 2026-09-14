package models

// Order структура для хранения информации о заказе.
type Order struct {
	Number      string  `json:"number"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual,omitempty"`
	UploadedAt  string  `json:"uploaded_at,omitempty"`
	ProcessedAt string  `json:"processed_at,omitempty"`
}
