package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krzko/oteldemo/internal/config"
	"github.com/krzko/oteldemo/pkg/simulator"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) == 1 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.Parse()

	if cfg.ShowVersion {
		printVersion()
		os.Exit(0)
	}

	sim, err := simulator.New(cfg)
	if err != nil {
		cfg.Logger.Error("Failed to create simulator", "error", err)
		os.Exit(1)
	}

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		cfg.Logger.Info("Received shutdown signal. Initiating graceful shutdown...")
		cancel()
	}()

	// Run the simulator
	go func() {
		if err := sim.Run(ctx); err != nil {
			cfg.Logger.Error("Simulator failed", "error", err)
			cancel()
		}
	}()

	// Wait for the context to be canceled
	<-ctx.Done()

	// Perform graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := sim.Shutdown(shutdownCtx); err != nil {
		cfg.Logger.Error("Error during shutdown", "error", err)
	}

	cfg.Logger.Info("Simulator shutdown complete")
}

func printUsage() {
	fmt.Printf("\nNAME:\n")
	fmt.Printf("  oteldemo - OpenTelemetry Demo Generator\n")

	fmt.Printf("\nDESCRIPTION:\n")
	fmt.Printf("  A tool to generate synthetic OpenTelemetry traces, metrics, and logs\n")
	fmt.Printf("  for testing and demonstration purposes.\n")

	fmt.Printf("\nVERSION:\n")
	fmt.Printf("  %s-%s (%s)\n", version, commit, date)

	fmt.Printf("\nUSAGE:\n")
	fmt.Printf("  oteldemo [flags]\n")
	fmt.Printf("  oteldemo -version\n")
	fmt.Printf("  oteldemo -help\n")

	fmt.Printf("\nFLAGS AND ENVIRONMENT VARIABLES:\n")
	fmt.Printf("  Flags can be set via command line or environment variables.\n")
	fmt.Printf("  Command line flags take precedence over environment variables.\n\n")
	fmt.Printf("  --endpoint, OTEL_EXPORTER_OTLP_ENDPOINT  (default: \"localhost:4318\")\n")
	fmt.Printf("    OpenTelemetry collector endpoint\n\n")
	fmt.Printf("  --protocol, OTEL_EXPORTER_OTLP_PROTOCOL  (default: \"http\")\n")
	fmt.Printf("    Protocol to use (grpc or http)\n\n")
	fmt.Printf("  --secure, OTEL_EXPORTER_OTLP_SECURE     (default: false)\n")
	fmt.Printf("    Use secure connection\n\n")
	fmt.Printf("  --headers, OTEL_EXPORTER_OTLP_HEADERS\n")
	fmt.Printf("    Headers as key=value pairs, comma-separated\n\n")
	fmt.Printf("  --env, APP_ENV                          (default: \"development\")\n")
	fmt.Printf("    Application environment\n\n")
	fmt.Printf("  -v, --version\n")
	fmt.Printf("    Display version information\n")

	fmt.Printf("\nEXAMPLES:\n")
	fmt.Printf("  # Local collector with HTTP:\n")
	fmt.Printf("  oteldemo -endpoint=\"localhost:4318\" -protocol=\"http\" -secure=\"false\"\n\n")
	fmt.Printf("  # Remote collector with gRPC and authentication:\n")
	fmt.Printf("  oteldemo -endpoint=\"api.domain.tld:443\" -headers=\"x-api-key=xxx,x-team=xxx\" -protocol=\"grpc\" -secure=\"true\"\n\n")
	fmt.Printf("  # Using environment variables:\n")
	fmt.Printf("  export OTEL_EXPORTER_OTLP_ENDPOINT=\"localhost:4318\"\n")
	fmt.Printf("  export OTEL_EXPORTER_OTLP_PROTOCOL=\"http\"\n")
	fmt.Printf("  oteldemo\n")

	fmt.Printf("\nDOCUMENTATION:\n")
	fmt.Printf("  https://github.com/krzko/oteldemo\n")
}

func printVersion() {
	fmt.Printf("oteldemo version: %s\n", version)
	fmt.Printf("git commit: %s\n", commit)
	fmt.Printf("built at: %s\n", date)
}
