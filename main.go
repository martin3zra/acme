package main

import (
	"embed"
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

func main() {

	err := godotenv.Load(os.ExpandEnv(".env"))
	if err != nil {
		panic(err)
	}

	server := app.NewServer(assets, resources)
	server.Boot()

	defer func() {
		server.Shutdown()
	}()

	go func() {
		server.Start()
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit
}
