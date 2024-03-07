package utils

import "context"

func ContextFn(ctx context.Context, fn func()) error {
	done := make(chan struct{})
	go func() {
		fn()
		done <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func ContextFnE(ctx context.Context, fn func() error) error {
	done := make(chan error)
	go func() {
		done <- fn()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
