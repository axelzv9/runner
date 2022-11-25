package runner

import (
	"time"
)

type Option func(*Runner)

// WithShutdownTimeout sets up shutdown timeout
func WithShutdownTimeout(shutdownTimeout time.Duration) Option {
	return func(runner *Runner) {
		if shutdownTimeout > 0 {
			runner.shutdownTimeout = shutdownTimeout
		}
	}
}

// WithShutdown adds graceful shutdown functions
func WithShutdown(fns ...Func) Option {
	return func(runner *Runner) {
		for _, fn := range fns {
			if fn == nil {
				continue
			}
			runner.shutdown = append(runner.shutdown, fn)
		}
	}
}
