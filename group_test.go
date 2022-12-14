package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	errTask := errors.New("failed task")

	testcases := []struct {
		name  string
		tasks []Func
		errs  []error
	}{
		{
			name: "basic case",
			tasks: []Func{
				func(ctx context.Context) error { return nil },
				func(ctx context.Context) error { return nil },
			},
			errs: nil,
		},
		{
			name: "task failed",
			tasks: []Func{
				func(ctx context.Context) error {
					return errTask
				},
				func(ctx context.Context) error {
					<-time.After(200 * time.Millisecond)
					return nil
				},
			},
			errs: []error{errTask},
		},
		{
			name: "task failed with timeout",
			tasks: []Func{
				func(ctx context.Context) error {
					<-time.After(200 * time.Millisecond)
					return errTask
				},
				func(ctx context.Context) error {
					return nil
				},
			},
			errs: []error{errTask},
		},
		{
			name: "both tasks failed",
			tasks: []Func{
				func(ctx context.Context) error {
					return errTask
				},
				func(ctx context.Context) error {
					return errTask
				},
			},
			errs: []error{errTask, errTask},
		},
		{
			name: "both tasks failed, one by timeout",
			tasks: []Func{
				func(ctx context.Context) error {
					return errTask
				},
				func(ctx context.Context) error {
					<-time.After(200 * time.Millisecond)
					return errTask
				},
			},
			errs: []error{errTask, errTask},
		},
	}

	t.Parallel()
	for _, testcase := range testcases {
		tc := testcase
		t.Run(tc.name, func(t *testing.T) {
			testGroup := NewErrorGroup(context.Background())

			testGroup.Go(tc.tasks...)

			err := testGroup.WaitFirst()
			if len(tc.errs) == 0 {
				assert.Nil(t, err)
			}
			if len(tc.errs) > 0 {
				assert.Equal(t, err, tc.errs[0])
			}

			errs := testGroup.WaitAll()
			assert.Equal(t, errs, tc.errs)
		})
	}
}
