package runner

import (
	"context"
	"sync"
	"testing"
)

const (
	benchmarkSizeSmall  = 10
	benchmarkSizeMedium = 1000
	benchmarkSizeLarge  = 100000
)

func Benchmark_Pool_Small(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workerPool(benchmarkSizeSmall)
	}
}

func Benchmark_Pool_Medium(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workerPool(benchmarkSizeMedium)
	}
}

func Benchmark_Pool_Large(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workerPool(benchmarkSizeLarge)
	}
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

func workerPool(size int) {
	wg := sync.WaitGroup{}
	wg.Add(size)

	tasks := make(chan func(), size)
	worker := func() {
		for task := range tasks {
			task()
		}
	}

	for j := 0; j < size; j++ {
		go worker()
		tasks <- func() {
			wg.Done()
		}
	}

	wg.Wait()
}

func errGroup(ctx context.Context, size int) {
	testGroup := NewErrorGroup(ctx)
	for j := 0; j < size; j++ {
		testGroup.Go(func(ctx context.Context) error {
			return nil
		})
	}
	testGroup.WaitFirst()
}

// goos: darwin
// goarch: arm64
// pkg: github.com/axelzv9/runner
// cpu: Apple M3 Pro
// Benchmark_Pool_Small
// Benchmark_Pool_Small-12         	  121042	     11866 ns/op	    6365 B/op	      33 allocs/op
// Benchmark_Pool_Medium
// Benchmark_Pool_Medium-12        	    1179	   1024814 ns/op	  621323 B/op	    3004 allocs/op
// Benchmark_Pool_Large
// Benchmark_Pool_Large-12         	       8	 126727276 ns/op	62227088 B/op	  300004 allocs/op
// Benchmark_ErrGroup_Small
// Benchmark_ErrGroup_Small-12     	  424855	      2886 ns/op	     544 B/op	      15 allocs/op
// Benchmark_ErrGroup_Medium
// Benchmark_ErrGroup_Medium-12    	    6507	    184624 ns/op	   24317 B/op	    1005 allocs/op
// Benchmark_ErrGroup_Large
// Benchmark_ErrGroup_Large-12     	      63	  18502929 ns/op	 2400327 B/op	  100005 allocs/op
