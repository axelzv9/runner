package runner

import (
	"context"
	"io"
)

func FuncWithError(fn func() error) Func {
	return func(_ context.Context) error {
		return fn()
	}
}

func FuncOnly(fn func()) Func {
	return func(_ context.Context) error {
		fn()
		return nil
	}
}

func Closer(closer io.Closer) Func {
	return func(_ context.Context) error {
		return closer.Close()
	}
}
