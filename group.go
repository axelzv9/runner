package runner

import (
	"context"
	"sync"
)

type ErrorGroup interface {
	Reset(ctx context.Context) ErrorGroup
	Go(fn ...Func) ErrorGroup
	WaitFirst() error
	WaitAll() []error
}

func NewErrorGroup(ctx context.Context) ErrorGroup {
	c, cancel := context.WithCancel(ctx)
	return &group{
		ctx:    c,
		cancel: cancel,
	}
}

type group struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errs errorSlice
}

func (g *group) Reset(ctx context.Context) ErrorGroup {
	g.ctx, g.cancel = context.WithCancel(ctx)
	g.errs.slice = g.errs.slice[0:]
	return g
}

func (g *group) Go(fn ...Func) ErrorGroup {
	for _, f := range fn {
		p := f
		g.wg.Add(1)
		go func() {
			err := p(g.ctx)
			g.done(err)
		}()
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
