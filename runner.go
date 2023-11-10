package runner

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Func func(context.Context) error
type InitFunc func(ctx context.Context, runner *Runner) error

const (
	defaultShutdownTimeout = 30 * time.Second
	defaultCleanTimeout    = 30 * time.Second
)

type Runner struct {
	shutdownTimeout time.Duration
	cleanupTimeout  time.Duration
	group           ErrorGroup

	mu       sync.Mutex
	shutdown []Func
	cleanup  []Func
}

func New(ctx context.Context, opts ...Option) *Runner {
	runner := &Runner{
		shutdownTimeout: defaultShutdownTimeout,
		cleanupTimeout:  defaultCleanTimeout,
		group:           NewErrorGroup(ctx),
	}

	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

func (r *Runner) Init(fn InitFunc) *Runner {
	return r.Run(func(ctx context.Context) error {
		return fn(ctx, r)
	})
}

func (r *Runner) Run(fn Func) *Runner {
	return r.RunGracefully(fn, nil)
}

func (r *Runner) RunGracefully(fn, shutdown Func) *Runner {
	r.AddShutdown(shutdown)
	r.group.Go(fn)
	return r
}

func (r *Runner) AddShutdown(shutdown ...Func) *Runner {
	for _, sd := range shutdown {
		if sd == nil {
			continue
		}
		r.mu.Lock()
		r.shutdown = append(r.shutdown, sd)
		r.mu.Unlock()
	}
	return r
}

func (r *Runner) AddCleanup(cleanup ...Func) *Runner {
	for _, fn := range cleanup {
		if fn == nil {
			continue
		}
		r.mu.Lock()
		r.cleanup = append(r.cleanup, fn)
		r.mu.Unlock()
	}
	return r
}

func (r *Runner) Wait() []error {
	// waiting for the first workers error
	_ = r.group.WaitFirst()

	errs := new(errorSlice)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), r.shutdownTimeout)
	defer shutdownCancel()

	_ = NewErrorGroup(shutdownCtx).Go(func(_ context.Context) error {
		// waiting for all workers
		errs.append(r.group.WaitAll()...)
		return nil
	}).Go(func(ctx context.Context) error {
		// waiting for shutdown process
		errs.append(NewErrorGroup(ctx).Go(r.shutdown...).WaitAll()...)
		return nil
	}).WaitFirst()

	if err := shutdownCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		errs.append(err)
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), r.cleanupTimeout)
	defer cleanupCancel()

	errs.append(NewErrorGroup(cleanupCtx).Go(r.cleanup...).WaitAll()...)

	return errs.items()
}
