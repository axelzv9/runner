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

	errs := runner.New(ctx, runner.WithShutdownTimeout(10*time.Second)).
		// init and run your jobs
		Init(func(ctx context.Context, group *runner.Runner) error {
			// add graceful shutdown for something
			group.AddShutdown(someShutdownFunc)

			// load config
			serverAddr := ":8080"

			// and init your dependencies here
			server := &http.Server{
				Addr: serverAddr,
			}

			dbConn := initDB()
			closeDBConn := func(ctx context.Context) error {
				return dbConn.Close()
			}
			// add graceful shutdown for database connection
			group.AddShutdown(closeDBConn)

			// run your jobs
			group.
				RunGracefully(runner.HTTPServer(server)).
				Run(someJob)

			return nil
		}).
		// waiting for results
		Wait()
	for _, err := range errs {
		log.Println("error occurred:", err)
	}
}

func someShutdownFunc(_ context.Context) error {
	log.Println("I'm shutting down...")
	return nil
}

func someJob(_ context.Context) error {
	time.Sleep(5 * time.Second)
	// job finished before http.Server, it does not affect http.Server, if it has no errors
	log.Println("someJob has been completed")
	return nil
}

type db struct{}

func initDB() *db {
	return &db{}
}

func (*db) Close() error {
	return nil
}
