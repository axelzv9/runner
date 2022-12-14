package runner

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestHTTPServer(t *testing.T) {
	httpServer := &http.Server{
		Addr: ":9090",
	}

	timeout := time.Second
	now := time.Now()

	New(context.Background()).
		RunGracefully(HTTPServer(httpServer)).
		Run(func(ctx context.Context) error {
			<-time.After(timeout)
			return errors.New("")
		}).
		Wait()

	if time.Since(now.Add(50*time.Millisecond)) > timeout {
		t.Fatal("too match time")
	}
}
