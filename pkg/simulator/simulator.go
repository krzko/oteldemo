package simulator

import (
	"context"
	"sync"

	"log/slog"

	"github.com/krzko/oteldemo/internal/config"
	"github.com/krzko/oteldemo/internal/services"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Simulator struct {
	services []services.Service
	logger   *slog.Logger
	tracer   trace.Tracer
}

func New(cfg *config.Config) (*Simulator, error) {
	sim := &Simulator{
		logger: cfg.Logger.With("component", "simulator"),
		tracer: otel.Tracer("simulator"),
	}
	for _, serviceName := range cfg.ServiceList {
		service, err := services.NewService(serviceName, cfg)
		if err != nil {
			sim.logger.Error("Failed to create service", "service", serviceName, "error", err)
			return nil, err
		}
		sim.services = append(sim.services, service)
	}
	return sim, nil
}

func (s *Simulator) Run(ctx context.Context) error {
	ctx, rootSpan := s.tracer.Start(ctx, "simulation")
	defer rootSpan.End()

	var wg sync.WaitGroup

	for _, service := range s.services {
		wg.Add(1)
		go func(svc services.Service) {
			defer wg.Done()
			svc.Simulate(ctx)
		}(service)
	}

	s.logger.Info("All services started")

	// Wait for all services to complete or context to be canceled
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("All services completed")
		return nil
	case <-ctx.Done():
		s.logger.Info("Simulation stopped by context cancellation")
		return ctx.Err()
	}
}

func (s *Simulator) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down simulator")
	var err error
	for _, service := range s.services {
		if shutdownErr := service.Shutdown(ctx); shutdownErr != nil {
			s.logger.Error("Error shutting down service", "error", shutdownErr)
			err = shutdownErr
		}
	}
	return err
}
