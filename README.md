# Golang task runner

Helper for running background tasks and managing their errors

## Usage example

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/axelzv9/runner"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	server := &http.Server{
		Addr: ":8080",
	}

	someJob := func(ctx context.Context) error {
		time.Sleep(5 * time.Second)
		// job finished before http.Server, it does not affect http.Server, if it has no errors
		log.Println("someJob has been completed")
		return nil
	}

	errs := runner.New(ctx).
		RunGracefully(runner.HTTPServer(server)).
		Run(someJob).
		Wait()
	for _, err := range errs {
		log.Println("error occurred:", err)
	}
}

```

See full integration example [here](https://github.com/axelzv9/runner/blob/main/example).