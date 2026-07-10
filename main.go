package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/martin3zra/acme/app"
)

//go:embed public/build
var assets embed.FS

//go:embed resources/views/*
var resources embed.FS

const (
	// exitFail is the exit code if the program
	// fails.
	exitFail = 1
)

func main() {

	err := godotenv.Load(os.ExpandEnv(".env"))
	if err != nil {
		panic(err)
	}
	err = os.Setenv("TZ", "America/Santo_Domingo")
	if err != nil {
		log.Fatalf("Error setting TZ env %v\n", err)
	}

	if err := run(os.Args, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)

		os.Exit(exitFail)
	}
}

func run(args []string, stdout io.Writer) error {
	app.InitLogger(stdout)

	server := app.NewServer(assets, resources)
	server.Boot()

	// Create a root context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler every 5 minutes
	server.StartScheduler(ctx, 10*time.Second)

	go func() {
		if err := server.Start(); err != nil {
			log.Printf("something wrong happens starting the server error: %v", err)
			cancel() // trigger shutdown
		}
	}()

	go func() {
		if err := server.StartSSE(); err != nil {
			log.Printf("something wrong happens starting the server error: %v", err)
			cancel() // trigger shutdown
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit
	log.Printf("shutting down gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	cancel() // stop scheduler

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	return nil
}
