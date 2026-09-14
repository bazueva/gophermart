package entities

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertValidatorErrors(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()

		result := ConvertValidatorErrors(nil)

		assert.Nil(t, result)
	})

	t.Run("error is not validator.ValidationErrors", func(t *testing.T) {
		t.Parallel()

		err := errors.New("some error")

		result := ConvertValidatorErrors(err)

		assert.Empty(t, result)
	})

	t.Run("required validation error", func(t *testing.T) {
		t.Parallel()

		validate := validator.New()

		type TestStruct struct {
			Login string `validate:"required"`
		}

		err := validate.Struct(TestStruct{})

		result := ConvertValidatorErrors(err)

		require.Len(t, result, 1)
		assert.Equal(t, ValidateError{
			Field: "login",
			Error: "Поле обязательно для заполнения",
		}, result[0])
	})

	t.Run("min validation error", func(t *testing.T) {
		t.Parallel()

		validate := validator.New()

		type TestStruct struct {
			Login string `validate:"min=5"`
		}

		err := validate.Struct(TestStruct{
			Login: "abc",
		})

		result := ConvertValidatorErrors(err)

		require.Len(t, result, 1)
		assert.Equal(t, ValidateError{
			Field: "login",
			Error: "Поле должно содержать не менее 5 символов",
		}, result[0])
	})

	t.Run("max validation error", func(t *testing.T) {
		t.Parallel()

		validate := validator.New()

		type TestStruct struct {
			Login string `validate:"max=5"`
		}

		err := validate.Struct(TestStruct{
			Login: "abcdef",
		})

		result := ConvertValidatorErrors(err)

		require.Len(t, result, 1)
		assert.Equal(t, ValidateError{
			Field: "login",
			Error: "Поле должно содержать не более 5 символов",
		}, result[0])
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		t.Parallel()

		validate := validator.New()

		type TestStruct struct {
			Login    string `validate:"required,min=5,max=10"`
			Password string `validate:"required"`
		}

		err := validate.Struct(TestStruct{
			Login:    "",
			Password: "",
		})

		result := ConvertValidatorErrors(err)

		require.Len(t, result, 2)

		assert.Equal(t, ValidateError{
			Field: "login",
			Error: "Поле обязательно для заполнения",
		}, result[0])

		assert.Equal(t, ValidateError{
			Field: "password",
			Error: "Поле обязательно для заполнения",
		}, result[1])
	})

	t.Run("field name converted to lowercase", func(t *testing.T) {
		t.Parallel()

		validate := validator.New()

		type TestStruct struct {
			UserLogin string `validate:"required"`
		}

		err := validate.Struct(TestStruct{})

		result := ConvertValidatorErrors(err)

		require.Len(t, result, 1)
		assert.Equal(t, "userlogin", result[0].Field)
	})
}
