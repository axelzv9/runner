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

func TestGroupWithPool(t *testing.T) {
	poolSize := int64(0)
	poolBufferSize := int64(0)
	for _, tc := range groupTestCases {
		poolSize = maxInt64(poolSize, int64(len(tc.tasks)))
	}
	poolSize = poolSize * int64(len(groupTestCases))
	poolBufferSize = poolSize * 2

	pool := NewPool(poolSize, poolBufferSize)

	t.Parallel()
	for _, tc := range groupTestCases {
		t.Run(tc.name, func(t *testing.T) {
			testGroup := NewErrorGroup(context.Background(), WithPool(pool))

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

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
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

func Benchmark_ErrGroupWithPool_Small(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()
	pool := NewPool(benchmarkSizeSmall, benchmarkSizeSmall)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errGroupWithPool(ctx, benchmarkSizeSmall, pool)
	}
}

func Benchmark_ErrGroupWithPool_Medium(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()
	pool := NewPool(benchmarkSizeMedium, benchmarkSizeMedium)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errGroupWithPool(ctx, benchmarkSizeMedium, pool)
	}
}

func Benchmark_ErrGroupWithPool_Large(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()
	pool := NewPool(benchmarkSizeLarge, benchmarkSizeLarge)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		errGroupWithPool(ctx, benchmarkSizeLarge, pool)
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

func errGroupWithPool(ctx context.Context, size int, pool *Pool) {
	testGroup := NewErrorGroup(ctx, WithPool(pool))
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
// Benchmark_ErrGroup_Small-12             	  363297	      3300 ns/op	     784 B/op	      35 allocs/op
// Benchmark_ErrGroup_Medium
// Benchmark_ErrGroup_Medium-12            	    5104	    229492 ns/op	   48352 B/op	    3005 allocs/op
// Benchmark_ErrGroup_Large
// Benchmark_ErrGroup_Large-12             	      54	  22929874 ns/op	 4800972 B/op	  300006 allocs/op
// Benchmark_ErrGroupWithPool_Small
// Benchmark_ErrGroupWithPool_Small-12     	  172608	      7153 ns/op	     624 B/op	      25 allocs/op
// Benchmark_ErrGroupWithPool_Medium
// Benchmark_ErrGroupWithPool_Medium-12    	    1009	   1099973 ns/op	   32492 B/op	    2006 allocs/op
// Benchmark_ErrGroupWithPool_Large
// Benchmark_ErrGroupWithPool_Large-12     	       6	 517663160 ns/op	 3456272 B/op	  202669 allocs/op
