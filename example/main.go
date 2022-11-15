package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/axelzv9/runner"
)

func main() {
	server := &http.Server{
		Addr: ":8080",
	}

	someJob := func(ctx context.Context) error {
		time.Sleep(5 * time.Second)
		// job finished before http.Server, it does not affect http.Server
		return nil
	}

	errs := runner.New(context.Background(), runner.WithSignalHandler()).
		RunGracefully(runner.HTTPServer(server)).
		Run(someJob).
		Wait()
	for _, err := range errs {
		log.Println("error occurred:", err)
	}
}
