package forms

// UserForm структура для формы регистрации пользователя.
type UserForm struct {
	Login    string `validate:"required,min=4,max=20"`
	Password string `validate:"required,min=8,max=32"`
}

// LoginForm структура для формы авторизации пользователя.
type LoginForm struct {
	Login    string `validate:"required,min=4,max=20"`
	Password string `validate:"required,min=8,max=32"`
}
