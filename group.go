package runner

import (
	"context"
	"sync"
)

type ErrorGroup interface {
	Go(fn ...Func) ErrorGroup
	WaitFirst() error
	WaitAll() []error
}

func NewErrorGroup(ctx context.Context, opts ...GroupOptions) ErrorGroup {
	c, cancel := context.WithCancel(ctx)
	g := &group{
		ctx:    c,
		cancel: cancel,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(g)
	}

	return g
}

type group struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	pool *Pool

	errs errorSlice
}

func (g *group) Go(fn ...Func) ErrorGroup {
	g.wg.Add(len(fn))

	if g.pool == nil {
		for _, f := range fn {
			go func(f Func) {
				g.done(f(g.ctx))
			}(f)
		}
		return g
	}

	for _, f := range fn {
		err := g.pool.Go(func(f Func) func() {
			return func() {
				g.done(f(g.ctx))
			}
		}(f))
		if err != nil {
			g.done(err)
		}
	}

	return g
}

func (g *group) WaitFirst() error {
	g.waitCtx()

	var err error
	if items := g.errs.items(); len(items) > 0 {
		err = items[0]
	}
	return err
}

func (g *group) WaitAll() []error {
	g.waitCtx()

	g.wg.Wait()

	return append([]error(nil), g.errs.items()...)
}

func (g *group) waitCtx() {
	go func() {
		g.wg.Wait()
		g.cancel()
	}()
	<-g.ctx.Done()
}

func (g *group) done(err error) {
	if err != nil {
		g.errs.append(err)
		g.cancel()
	}
	g.wg.Done()
}

type errorSlice struct {
	mu    sync.Mutex
	slice []error
}

func (e *errorSlice) append(errs ...error) {
	e.mu.Lock()
	e.slice = append(e.slice, errs...)
	e.mu.Unlock()
}

func (e *errorSlice) items() []error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.slice
}
