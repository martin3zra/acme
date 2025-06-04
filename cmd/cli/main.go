package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/martin3zra/acme/app"
	"github.com/martin3zra/acme/pkg/foundation/str"
)

const (
	// exitFail is the exit code if the program
	// fails.
	exitFail = 1
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mycli <command> [arguments]")
		return
	}

	err := godotenv.Load(os.ExpandEnv("../../.env"))
	if err != nil {
		panic(err)
	}

	err = os.Setenv("TZ", "America/Santo_Domingo")
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

func run(args []string, stdout io.Writer) error {
	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile("../../acme.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// optional: log date-time, filename, and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.SetOutput(file)

	server := app.NewServer(nil, nil)
	server.Boot()

	defer func() {
		server.Shutdown()
		log.Println("Stopping the server")
	}()

	log.Println("Starting the server")

	menu(args, server)

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
		fmt.Println("mycli version 1.0.0")
	case "generate:key":
		fmt.Println("generate key", str.GenerateRandom())
	case "setup:account":
		server.SetupAccount()
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}
