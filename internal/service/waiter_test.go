package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaiter_Set(t *testing.T) {
	t.Parallel()

	t.Run("устанавливает время ожидания", func(t *testing.T) {
		t.Parallel()

		waiter := NewWaiter()

		now := time.Now()
		waiter.Set(time.Second)

		until := time.Unix(0, waiter.until.Load())

		assert.True(t, until.After(now))
		assert.True(t, until.Before(now.Add(2*time.Second)))
	})

	t.Run("не уменьшает уже установленное время ожидания", func(t *testing.T) {
		t.Parallel()

		waiter := NewWaiter()

		waiter.Set(2 * time.Second)
		firstUntil := waiter.until.Load()

		waiter.Set(time.Second)
		secondUntil := waiter.until.Load()

		assert.Equal(t, firstUntil, secondUntil)
	})

	t.Run("увеличивает время ожидания", func(t *testing.T) {
		t.Parallel()

		waiter := NewWaiter()

		waiter.Set(time.Second)
		firstUntil := waiter.until.Load()

		waiter.Set(2 * time.Second)
		secondUntil := waiter.until.Load()

		assert.Greater(t, secondUntil, firstUntil)
	})
}

func TestWaiter_Wait(t *testing.T) {
	t.Parallel()

	t.Run("не ждет если время ожидания истекло", func(t *testing.T) {
		t.Parallel()

		waiter := NewWaiter()

		start := time.Now()

		err := waiter.Wait(t.Context())

		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, elapsed, 100*time.Millisecond)
	})

	t.Run("ждет установленное время", func(t *testing.T) {
		t.Parallel()

		waiter := NewWaiter()
		waiter.Set(100 * time.Millisecond)

		start := time.Now()

		err := waiter.Wait(t.Context())

		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, elapsed, 80*time.Millisecond)
	})

	t.Run("прерывает ожидание при отмене контекста", func(t *testing.T) {
		t.Parallel()

		waiter := NewWaiter()
		waiter.Set(time.Second)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := waiter.Wait(ctx)

		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}