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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	errs := runner.New(ctx).Init(func(ctx context.Context, group *runner.Runner) error {
		server := &http.Server{
			Addr: ":8080",
		}

		someJob := func(ctx context.Context) error {
			time.Sleep(5 * time.Second)
			// job finished before http.Server, it does not affect http.Server, if it has no errors
			log.Println("someJob has been completed")
			return nil
		}

		group.RunGracefully(runner.HTTPServer(server)).Run(someJob)
		return nil
	}).Wait()
	for _, err := range errs {
		log.Println("error occurred:", err)
	}
}
