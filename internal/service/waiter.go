package service

import (
	"context"
	"sync/atomic"
	"time"
)

type waiter struct {
	until atomic.Int64
}

func NewWaiter() *waiter {
	return &waiter{}
}

func (w *waiter) Set(duration time.Duration) {
	newUntil := time.Now().Add(duration).UnixNano()

	if newUntil <= w.until.Load() {
		return
	}

	w.until.Store(newUntil)
}

func (w *waiter) Wait(ctx context.Context) error {
	until := time.Unix(0, w.until.Load())

	if !until.After(time.Now()) {
		return nil
	}

	timer := time.NewTimer(time.Until(until))
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}