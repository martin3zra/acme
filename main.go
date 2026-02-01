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
	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile("acme.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// optional: log date-time, filename, and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.SetOutput(file)

	server := app.NewServer(assets, resources)
	server.Boot()

	// Create a root context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler every 5 minutes
	server.StartScheduler(ctx, 5*time.Minute)

	go func() {
		if err := server.Start(); err != nil {
			log.Printf("server error %v", err)
			cancel() // trigger shutdown
		}
		log.Println("Starting the server")
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
