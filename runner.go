package runner

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Func func(context.Context) error

const defaultShutdownTimeout = 30 * time.Second

type Runner struct {
	ctx             context.Context
	cancel          context.CancelFunc
	shutdownTimeout time.Duration
	group           ErrorGroup

	mu       sync.Mutex
	shutdown []Func
}

func New(opts ...Option) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	runner := &Runner{
		ctx:             ctx,
		cancel:          cancel,
		shutdownTimeout: defaultShutdownTimeout,
		group:           NewErrorGroup(ctx),
	}

	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

func (r *Runner) Run(fn Func) *Runner {
	return r.RunGracefully(fn, nil)
}

func (r *Runner) RunGracefully(fn, shutdown Func) *Runner {
	if shutdown != nil {
		r.mu.Lock()
		r.shutdown = append(r.shutdown, shutdown)
		r.mu.Unlock()
	}

	r.group.Go(fn)
	return r
}

func (r *Runner) Wait() []error {
	// waiting for the first workers error
	_ = r.group.WaitFirst()

	errs := new(Errors)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), r.shutdownTimeout)
	defer shutdownCancel()

	_ = NewErrorGroup(shutdownCtx).Go(func(_ context.Context) error {
		// waiting for all workers
		errs.Append(r.group.WaitAll()...)
		return nil
	}).Go(func(ctx context.Context) error {
		// waiting for shutdown process
		errs.Append(NewErrorGroup(ctx).Go(r.shutdown...).WaitAll()...)
		return nil
	}).WaitFirst()

	if err := shutdownCtx.Err(); err == nil || errors.Is(err, context.Canceled) {
		return errs.Errors()
	}
	return append(errs.Errors(), shutdownCtx.Err())
}

type Errors struct {
	mu   sync.Mutex
	errs []error
}

func (e *Errors) Append(errs ...error) {
	e.mu.Lock()
	e.errs = append(e.errs, errs...)
	e.mu.Unlock()
}

func (e *Errors) Errors() []error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.errs
}
