package runner

import (
	"os"
	"os/signal"
	"syscall"
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

// WithSignalHandler handles SIGINT and SIGTERM signals by default
func WithSignalHandler(signals ...os.Signal) Option {
	return func(runner *Runner) {
		signals = append(signals, syscall.SIGINT, syscall.SIGTERM)
		runner.ctx, _ = signal.NotifyContext(runner.ctx, signals...)
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
