package runner

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolTaskSubmission(t *testing.T) {
	p := NewPool(2, 1)

	err := p.Go(func() { time.Sleep(10 * time.Millisecond) })
	assert.NoError(t, err)

	err = p.Go(func() {})
	assert.ErrorIs(t, err, ErrOverloaded)

	p.Close()

	err = p.Go(func() {})
	assert.ErrorIs(t, err, ErrClosed)
}

func TestPoolRescale(t *testing.T) {
	p := NewPool(2, 1)

	p.Rescale(4)
	assert.Equal(t, int64(4), p.size.Load())

	p.Rescale(1)
	assert.Equal(t, int64(1), p.size.Load())
}

func TestPoolClose(t *testing.T) {
	p := NewPool(2, 2)

	done := make(chan struct{})
	err := p.Go(func() {
		time.Sleep(10 * time.Millisecond)
		close(done)
	})
	assert.NoError(t, err)

	p.Close()

	select {
	case <-done:
		// pass
	case <-time.After(50 * time.Millisecond):
		t.Error("task was not executed before pool closed")
	}

	err = p.Go(func() {})
	assert.ErrorIs(t, err, ErrClosed)
}

func TestPoolDownscale(t *testing.T) {
	p := NewPool(10, 10)

	var counter int64
	for i := 0; i < 10; i++ {
		err := p.Go(func() {
			atomic.AddInt64(&counter, 1)
			time.Sleep(10 * time.Millisecond)
		})
		assert.NoError(t, err)
	}

	p.Rescale(1)

	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, int64(10), atomic.LoadInt64(&counter))
}

func TestPoolMinSize(t *testing.T) {
	p := NewPool(10, 10)
	p.Rescale(0)
	assert.Equal(t, minPoolSize, p.size.Load())
}

func TestPoolHighLoad(t *testing.T) {
	p := NewPool(100, 100)

	var counter int64
	for i := 0; i < 1000; i++ {
		go func() {
			for {
				err := p.Go(func() {
					atomic.AddInt64(&counter, 1)
					time.Sleep(1 * time.Millisecond)
				})
				if err == nil {
					break
				}
				assert.ErrorIs(t, err, ErrOverloaded)
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int64(1000), atomic.LoadInt64(&counter))
}
