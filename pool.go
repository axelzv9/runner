package runner

import (
	"errors"
	"sync/atomic"
)

const (
	minPoolSize          = int64(1)
	minBufferSize        = int64(1)
	defaultDownscaleSize = int64(100)
)

var (
	ErrClosed     = errors.New("closed")
	ErrOverloaded = errors.New("overloaded")
)

type Pool struct {
	tasks     chan func()
	count     *atomic.Int64
	size      *atomic.Int64
	closed    *atomic.Bool
	downscale chan struct{}
}

func NewPool(size, bufferSize int64) *Pool {
	if size < minPoolSize {
		size = minPoolSize
	}
	if bufferSize < minBufferSize {
		bufferSize = minBufferSize
	}

	p := &Pool{
		tasks:     make(chan func(), bufferSize),
		count:     new(atomic.Int64),
		size:      new(atomic.Int64),
		closed:    new(atomic.Bool),
		downscale: make(chan struct{}, defaultDownscaleSize),
	}

	p.Rescale(size)

	return p
}

func (p *Pool) Go(fn ...func()) error {
	if p.closed.Load() || p.size.Load() < 1 {
		return ErrClosed
	}

	for _, f := range fn {
		select {
		case p.tasks <- f:
		default:
			return ErrOverloaded
		}
	}

	return nil
}

func (p *Pool) Rescale(size int64) {
	if size < minPoolSize {
		size = minPoolSize
	}

	p.size.Store(size)
	cnt := p.count.Load()

	// upscale
	for i := cnt; i < size; i++ {
		go p.worker()
	}

	// downscale
	for i := size; i < cnt; i++ {
		p.downscale <- struct{}{}
	}
}

func (p *Pool) Close() {
	p.closed.Store(true)
	close(p.tasks)
}

func (p *Pool) worker() {
	p.count.Add(1)
	defer p.count.Add(-1)

	for {
		select {
		case fn, ok := <-p.tasks:
			if !ok {
				return
			}
			fn()
		case <-p.downscale:
			return
		}
	}
}
