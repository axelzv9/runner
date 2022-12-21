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
