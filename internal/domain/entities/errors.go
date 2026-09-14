package entities

import (
	"github.com/samber/lo"
)

// DomainError структура для доменных ошибок.
type DomainError struct {
	ErrorType ErrorType
	SourceErr error
	Text      string
}

// ErrorType тип ошибки.
type ErrorType int

const (
	InternalServerErrorType ErrorType = iota
	ConflictErrorType
	BadRequestErrorType
	UnauthorizedErrorType
	UnprocessableEntityErrorType
	OkEntityErrorType
	NoContentErrorType
	RetriableErrorType
	PaymentRequiredErrorType
	TooManyRequestErrorType
)

// Error возвращает текст ошибки.
func (e *DomainError) Error() string {
	return e.Text
}

// NewInternalServerError создает ошибку Internal Server Error.
func NewInternalServerError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: InternalServerErrorType,
		SourceErr: err,
		Text:      lo.Ternary(text != "", text, "Internal Server Error"),
	}
}

// NewUnauthorizedError создает ошибку Unauthorized.
func NewUnauthorizedError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: UnauthorizedErrorType,
		SourceErr: err,
		Text:      lo.Ternary(text != "", text, "пользователь не аутентифицирован"),
	}
}

// NewOkError создает ошибку Ok.
func NewOkError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: OkEntityErrorType,
		SourceErr: err,
		Text:      text,
	}
}

// NewUnprocessableEntity создает ошибку UnprocessableEntity.
func NewUnprocessableEntity(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: UnprocessableEntityErrorType,
		SourceErr: err,
		Text:      text,
	}
}

func NewConflictError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: ConflictErrorType,
		SourceErr: err,
		Text:      text,
	}
}

// NewBadRequestError создает ошибку BadRequest.
func NewBadRequestError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: BadRequestErrorType,
		SourceErr: err,
		Text:      text,
	}
}

// NewNoContentError создает ошибку NoContent.
func NewNoContentError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: NoContentErrorType,
		SourceErr: err,
		Text:      text,
	}
}

// NewRetriableError создает ошибку Retriable.
func NewRetriableError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: RetriableErrorType,
		SourceErr: err,
		Text:      text,
	}
}

// NewPaymentRequiredError создает ошибку PaymentRequired.
func NewPaymentRequiredError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: PaymentRequiredErrorType,
		SourceErr: err,
		Text:      text,
	}
}

// NewTooManyRequestError создает ошибку TooManyRequest.
func NewTooManyRequestError(err error, text string) *DomainError {
	return &DomainError{
		ErrorType: TooManyRequestErrorType,
		SourceErr: err,
		Text:      lo.Ternary(text != "", text, "Слишком много запросов"),
	}
}
