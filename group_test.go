package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	benchmarkSizeSmall  = 10
	benchmarkSizeMedium = 1000
	benchmarkSizeLarge  = 100000
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

func Benchmark_ErrGroup_Small(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errGroup(ctx, benchmarkSizeSmall)
	}
}

func Benchmark_ErrGroup_Medium(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errGroup(ctx, benchmarkSizeMedium)
	}
}

func Benchmark_ErrGroup_Large(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errGroup(ctx, benchmarkSizeLarge)
	}
}

var benchGroupSmall = NewErrorGroup(context.Background())

func Benchmark_ErrGroupWithReset_Small(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errGroupWithReset(ctx, benchmarkSizeSmall, benchGroupSmall)
	}
}

var benchGroupMedium = NewErrorGroup(context.Background())

func Benchmark_ErrGroupWithReset_Medium(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errGroupWithReset(ctx, benchmarkSizeMedium, benchGroupMedium)
	}
}

var benchGroupLarge = NewErrorGroup(context.Background())

func Benchmark_ErrGroupWithReset_Large(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errGroupWithReset(ctx, benchmarkSizeLarge, benchGroupLarge)
	}
}

func errGroup(ctx context.Context, size int) {
	testGroup := NewErrorGroup(ctx)
	for i := 0; i < size; i++ {
		testGroup.Go(func(ctx context.Context) error {
			return nil
		})
	}
	_ = testGroup.WaitFirst()
}

func errGroupWithReset(ctx context.Context, size int, errGroup ErrorGroup) {
	testGroup := errGroup.Reset(ctx)
	for i := 0; i < size; i++ {
		testGroup.Go(func(ctx context.Context) error {
			return nil
		})
	}
	_ = testGroup.WaitFirst()
}

// goos: darwin
// goarch: arm64
// pkg: github.com/axelzv9/runner
// cpu: Apple M3 Pro
// Benchmark_ErrGroup_Small
// Benchmark_ErrGroup_Small-12              	  440301	      2925 ns/op	     544 B/op	      15 allocs/op
// Benchmark_ErrGroup_Medium
// Benchmark_ErrGroup_Medium-12             	    6710	    178567 ns/op	   24339 B/op	    1005 allocs/op
// Benchmark_ErrGroup_Large
// Benchmark_ErrGroup_Large-12              	      67	  18120942 ns/op	 2411259 B/op	  100027 allocs/op
// Benchmark_ErrGroupWithReset_Small
// Benchmark_ErrGroupWithReset_Small-12     	  373756	      3169 ns/op	     544 B/op	      24 allocs/op
// Benchmark_ErrGroupWithReset_Medium
// Benchmark_ErrGroupWithReset_Medium-12    	    6032	    196009 ns/op	   32224 B/op	    2004 allocs/op
// Benchmark_ErrGroupWithReset_Large
// Benchmark_ErrGroupWithReset_Large-12     	      61	  19519690 ns/op	 3200241 B/op	  200004 allocs/op
