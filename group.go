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

	mu   sync.Mutex
	errs []error
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
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.errs) > 0 {
		err = g.errs[0]
	}
	return err
}

func (g *group) WaitAll() []error {
	g.waitCtx()

	g.wg.Wait()

	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]error(nil), g.errs...)
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
		g.mu.Lock()
		g.errs = append(g.errs, err)
		g.mu.Unlock()

		g.cancel()
	}
	g.wg.Done()
}
