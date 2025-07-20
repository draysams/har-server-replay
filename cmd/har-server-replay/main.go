package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/draysams/har-server-replay/internal/har"
	"github.com/draysams/har-server-replay/internal/server"
)

// Run executes the main logic for the HAR replay server CLI.
func Run(args []string) error {
	flagSet := flag.NewFlagSet("har-server-replay", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)

	harFile := flagSet.String("har-file", "", "Path to the HAR file (required)")
	port := flagSet.Int("port", 8080, "Port to listen on")
	verbose := flagSet.Bool("verbose", false, "Enable verbose logging")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if *harFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --har-file is required")
		flagSet.Usage()
		return fmt.Errorf("--har-file is required")
	}
	server.SetVerbose(*verbose)

	harData, err := har.LoadAndParse(*harFile)
	if err != nil {
		return fmt.Errorf("Error loading and parsing HAR file: %w", err)
	}

	if err := server.Start(*port, harData); err != nil {
		return fmt.Errorf("Error starting server: %w", err)
	}
	return nil
}

func main() {
	if err := Run(os.Args[1:]); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
