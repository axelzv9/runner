package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var errTask = errors.New("failed task")

func TestGroup(t *testing.T) {
	t.Parallel()
	for _, tc := range groupTestCases {
		t.Run(tc.name, func(t *testing.T) {
			testGroup := NewErrorGroup(context.Background())

			testGroup.Go(tc.tasks...)

			err := testGroup.WaitFirst()
			if len(tc.errs) == 0 {
				assert.NoError(t, err)
			}
			if len(tc.errs) > 0 {
				assert.Equal(t, tc.errs[0], err)
			}

			errs := testGroup.WaitAll()
			assert.Equal(t, tc.errs, errs)
		})
	}
}

var groupTestCases = []struct {
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
