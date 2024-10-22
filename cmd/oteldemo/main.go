package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krzko/oteldemo/internal/config"
	"github.com/krzko/oteldemo/pkg/simulator"
)

func main() {
	cfg := config.Parse()
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
