package models

// RegisterRequest структура для запроса на регистрацию.
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginRequest структура для запроса на авторизацию.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// BalanceWithdrawRequest структура для запроса на списание бонусов.
type BalanceWithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
