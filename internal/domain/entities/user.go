package entities

// User структура пользователя.
type User struct {
	ID           int32
	Login        string
	PasswordHash string
}
