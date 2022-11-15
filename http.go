package runner

import (
	"context"
	"errors"
	"net/http"
)

// HTTPServer is wrapper function for launch http server.
func HTTPServer(srv *http.Server) (fn, shutdown Func) {
	return func(ctx context.Context) error {
			err := srv.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return err
		}, func(ctx context.Context) error {
			err := srv.Shutdown(ctx)
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return err
		}
}
