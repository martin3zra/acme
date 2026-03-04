package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/martin3zra/acme/app"
	"github.com/martin3zra/acme/pkg/foundation/str"
	"github.com/martin3zra/acme/resources"
)

const (
	// exitFail is the exit code if the program
	// fails.
	exitFail = 1
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: Acme CLI <command> [arguments]")
		return
	}

	loadEnv()

	err := os.Setenv("TZ", "America/Santo_Domingo")
	if err != nil {
		log.Fatalf("Error setting TZ env %v\n", err)
	}

	err = os.Setenv("RUNNING_IN_CLI", "YES")
	if err != nil {
		log.Fatalf("Error setting RUNNING_IN_CLI env %v\n", err)
	}

	if err := run(os.Args, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)

		os.Exit(exitFail)
	}

}

func loadEnv() {
	// Try current working directory
	wd, _ := os.Getwd()
	envPath := filepath.Join(wd, ".env")

	if err := godotenv.Load(envPath); err == nil {
		return
	}

	// Fallback: relative to binary (useful in production builds)
	exePath, _ := os.Executable()
	rootPath := filepath.Join(filepath.Dir(exePath), "..", "..")
	envPath = filepath.Join(rootPath, ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
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

	var assets embed.FS
	server := app.NewServer(assets, resources.Views)
	server.Boot()

	// Create a root context with cancel
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	log.Println("Starting the server")

	menu(args, server)

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	return nil
}

func menu(args []string, server *app.Server) {

	// Load environment and set we're as CLI env
	// Do not load any UI stuff like view, routes, etc.
	command := args[1]
	args = args[2:]

	switch command {
	case "greet":
		if len(args) > 0 {
			fmt.Printf("Hello, %s!\n", args[0])
		} else {
			fmt.Println("Hello, world!")
		}
	case "version":
		fmt.Println("Acme CLI version 1.0.0")
	case "generate:key":
		fmt.Println("generate key", str.GenerateRandom())
	case "setup:account":
		server.SetupAccount()
	case "resend:account-email-verification":
		server.ResendAccountVerificationEmail()
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
