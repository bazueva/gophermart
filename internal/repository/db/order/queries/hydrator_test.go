package queries

import (
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	"github.com/bazueva/gofermart/schema.gen/gofermart/public/enum"
	dbModel "github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/stretchr/testify/assert"
)

func Test_hydrateDomainToOrdersStatusEnum(t *testing.T) {
	t.Parallel()

	t.Run("convert OrdersStatusNew to enum", func(t *testing.T) {
		status := entities.OrdersStatusNew
		result := hydrateDomainToOrdersStatusEnum(status)

		assert.Equal(t, enum.OrdersStatus.New, result)
	})

	t.Run("convert OrdersStatusProcessing to enum", func(t *testing.T) {
		status := entities.OrdersStatusProcessing
		result := hydrateDomainToOrdersStatusEnum(status)

		assert.Equal(t, enum.OrdersStatus.Processing, result)
	})

	t.Run("convert OrdersStatusProcessed to enum", func(t *testing.T) {
		status := entities.OrdersStatusProcessed
		result := hydrateDomainToOrdersStatusEnum(status)

		assert.Equal(t, enum.OrdersStatus.Processed, result)
	})

	t.Run("convert OrdersStatusInvalid to enum", func(t *testing.T) {
		status := entities.OrdersStatusInvalid
		result := hydrateDomainToOrdersStatusEnum(status)

		assert.Equal(t, enum.OrdersStatus.Invalid, result)
	})

	t.Run("convert unknown status returns default (New)", func(t *testing.T) {
		status := entities.OrderStatus("UNKNOWN")
		result := hydrateDomainToOrdersStatusEnum(status)

		assert.Equal(t, enum.OrdersStatus.New, result)
	})

	t.Run("convert empty status returns default (New)", func(t *testing.T) {
		status := entities.OrderStatus("")
		result := hydrateDomainToOrdersStatusEnum(status)

		assert.Equal(t, enum.OrdersStatus.New, result)
	})
}

func Test_hydrateDomainToOrdersStatus(t *testing.T) {
	t.Parallel()

	t.Run("convert OrdersStatusNew to db model", func(t *testing.T) {
		status := entities.OrdersStatusNew
		result := hydrateDomainToOrdersStatus(status)

		assert.Equal(t, dbModel.OrdersStatus_New, result)
	})

	t.Run("convert OrdersStatusProcessing to db model", func(t *testing.T) {
		status := entities.OrdersStatusProcessing
		result := hydrateDomainToOrdersStatus(status)

		assert.Equal(t, dbModel.OrdersStatus_Processing, result)
	})

	t.Run("convert OrdersStatusProcessed to db model", func(t *testing.T) {
		status := entities.OrdersStatusProcessed
		result := hydrateDomainToOrdersStatus(status)

		assert.Equal(t, dbModel.OrdersStatus_Processed, result)
	})

	t.Run("convert OrdersStatusInvalid to db model", func(t *testing.T) {
		status := entities.OrdersStatusInvalid
		result := hydrateDomainToOrdersStatus(status)

		assert.Equal(t, dbModel.OrdersStatus_Invalid, result)
	})

	t.Run("convert unknown status returns default", func(t *testing.T) {
		status := entities.OrderStatus("UNKNOWN")
		result := hydrateDomainToOrdersStatus(status)

		assert.Equal(t, dbModel.OrdersStatus("UNKNOWN"), result)
	})

	t.Run("convert empty status returns default", func(t *testing.T) {
		status := entities.OrderStatus("")
		result := hydrateDomainToOrdersStatus(status)

		assert.Equal(t, dbModel.OrdersStatus(""), result)
	})
}
