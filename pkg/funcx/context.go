package funcx

import "context"

func Context(ctx context.Context, fn func()) error {
	return WrapContext(ctx, func(_ context.Context) { fn() })
}

func ContextE(ctx context.Context, fn func() error) error {
	return WrapContextE(ctx, func(_ context.Context) error { return fn() })
}

func WrapContext(ctx context.Context, fn func(ctx context.Context)) error {
	done := make(chan struct{})
	go func() {
		fn(ctx)
		done <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func WrapContextE(ctx context.Context, fn func(context.Context) error) error {
	done := make(chan error)
	go func() {
		done <- fn(ctx)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
