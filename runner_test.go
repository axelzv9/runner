package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunner(t *testing.T) {
	errTask := errors.New("failed task")
	errShutdown := errors.New("shutdown failed")
	errShutdown2 := errors.New("shutdown failed 2")

	testcases := []struct {
		name            string
		background      [2]Func
		shutdown        []Func
		shutdownTimeout time.Duration
		errs            []error
	}{
		{
			name: "basic case",
			background: [2]Func{
				func(ctx context.Context) error { return nil },
				func(ctx context.Context) error { return nil },
			},
			errs: nil,
		},
		{
			name: "background failed",
			background: [2]Func{
				func(ctx context.Context) error {
					<-time.After(200 * time.Millisecond)
					return errTask
				},
				func(ctx context.Context) error { return nil },
			},
			errs: []error{errTask},
		},
		{
			name: "shutdown failed",
			background: [2]Func{
				func(ctx context.Context) error {
					<-time.After(200 * time.Millisecond)
					return nil
				},
				func(ctx context.Context) error {
					return errShutdown
				},
			},
			errs: []error{errShutdown},
		},
		{
			name: "background and shutdown failed",
			background: [2]Func{
				func(ctx context.Context) error {
					<-time.After(time.Millisecond)
					return errTask
				},
				func(ctx context.Context) error {
					<-time.After(2 * time.Millisecond)
					return errShutdown
				},
			},
			errs: []error{errTask, errShutdown},
		},
		{
			name: "background failed and shutdown failed by timeout",
			background: [2]Func{
				func(ctx context.Context) error {
					return errTask
				},
				func(ctx context.Context) error {
					<-time.After(200 * time.Millisecond)
					return errShutdown
				},
			},
			errs: []error{errTask, errShutdown},
		},
		{
			name: "background failed and shutdown failed by timeout 2",
			background: [2]Func{
				func(ctx context.Context) error {
					return errTask
				},
				func(ctx context.Context) error {
					<-time.After(2 * time.Second)
					return errShutdown
				},
			},
			shutdownTimeout: time.Second,
			errs:            []error{errTask, context.DeadlineExceeded},
		},
		{
			name: "shutdown failed",
			background: [2]Func{
				func(ctx context.Context) error { return nil },
				func(ctx context.Context) error { return nil },
			},
			shutdown: []Func{
				func(ctx context.Context) error { return errShutdown },
			},
			errs: []error{errShutdown},
		},
		{
			name: "shutdown failed by timeout",
			background: [2]Func{
				func(ctx context.Context) error { return nil },
				func(ctx context.Context) error { return nil },
			},
			shutdown: []Func{
				func(ctx context.Context) error {
					<-time.After(200 * time.Millisecond)
					return errShutdown
				},
			},
			errs: []error{errShutdown},
		},
		{
			name: "shutdown multiple failed",
			background: [2]Func{
				func(ctx context.Context) error { return nil },
				func(ctx context.Context) error { return nil },
			},
			shutdown: []Func{
				func(ctx context.Context) error {
					<-time.After(300 * time.Millisecond)
					return errShutdown
				},
				func(ctx context.Context) error {
					<-time.After(500 * time.Millisecond)
					return errShutdown2
				},
			},
			errs: []error{errShutdown, errShutdown2},
		},
	}

	t.Parallel()
	for _, testcase := range testcases {
		tc := testcase
		t.Run(tc.name, func(t *testing.T) {
			errs := New(context.Background(), WithShutdownTimeout(tc.shutdownTimeout)).
				Init(func(ctx context.Context, group *Runner) error {
					group.AddShutdown(tc.shutdown...)
					return nil
				}).
				RunGracefully(tc.background[0], tc.background[1]).
				Wait()

			assert.Equal(t, tc.errs, errs)
		})
	}
}
