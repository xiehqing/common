package safego

import "context"

// Go 安全的go, 捕获panic
func Go(ctx context.Context, f func()) {
	go func() {
		defer Recovery(ctx)
		f()
	}()
}
