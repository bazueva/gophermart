package order

import (
	"testing"

	"github.com/bazueva/gofermart/internal/domain/entities"
	dbModel "github.com/bazueva/gofermart/schema.gen/gofermart/public/model"
	"github.com/stretchr/testify/assert"
)

func Test_hydrateOrdersStatusToDomain(t *testing.T) {
	t.Parallel()

	t.Run("convert db OrdersStatus_New to domain", func(t *testing.T) {
		status := dbModel.OrdersStatus_New
		result := hydrateOrdersStatusToDomain(status)

		assert.Equal(t, entities.OrdersStatusNew, result)
	})

	t.Run("convert db OrdersStatus_Processing to domain", func(t *testing.T) {
		status := dbModel.OrdersStatus_Processing
		result := hydrateOrdersStatusToDomain(status)

		assert.Equal(t, entities.OrdersStatusProcessing, result)
	})

	t.Run("convert db OrdersStatus_Processed to domain", func(t *testing.T) {
		status := dbModel.OrdersStatus_Processed
		result := hydrateOrdersStatusToDomain(status)

		assert.Equal(t, entities.OrdersStatusProcessed, result)
	})

	t.Run("convert db OrdersStatus_Invalid to domain", func(t *testing.T) {
		status := dbModel.OrdersStatus_Invalid
		result := hydrateOrdersStatusToDomain(status)

		assert.Equal(t, entities.OrdersStatusInvalid, result)
	})

	t.Run("convert unknown db status returns default (New)", func(t *testing.T) {
		status := dbModel.OrdersStatus("UNKNOWN")
		result := hydrateOrdersStatusToDomain(status)

		assert.Equal(t, entities.OrdersStatusNew, result)
	})

	t.Run("convert empty db status returns default (New)", func(t *testing.T) {
		status := dbModel.OrdersStatus("")
		result := hydrateOrdersStatusToDomain(status)

		assert.Equal(t, entities.OrdersStatusNew, result)
	})
}
