package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	server := app.NewServer(&assets, &resources)
	server.Boot()

	defer func() {
		server.Shutdown()
		log.Println("Stopping the server")
	}()

	go func() {
		server.Start()
		log.Println("Starting the server")
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit

	return nil
}
