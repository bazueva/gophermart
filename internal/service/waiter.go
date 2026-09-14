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

	for {
		currentUntil := w.until.Load()

		if newUntil <= currentUntil {
			return
		}

		if w.until.CompareAndSwap(currentUntil, newUntil) {
			return
		}
	}
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