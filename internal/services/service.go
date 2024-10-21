package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/krzko/oteldemo/internal/config"
	"github.com/krzko/oteldemo/internal/telemetry"
	"github.com/krzko/oteldemo/pkg/data"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type Service interface {
	Simulate(ctx context.Context)
	Shutdown(ctx context.Context) error
}

type BaseService struct {
	Name      string
	Tracer    trace.Tracer
	Provider  *sdktrace.TracerProvider
	Generator *data.Generator
	Logger    *slog.Logger
	cfg       *config.Config
}

func NewService(name string, cfg *config.Config) (Service, error) {
	tp, err := telemetry.NewProvider(name, cfg.Endpoint, cfg.Secure, cfg.Protocol, cfg.Headers)
	if err != nil {
		return nil, err
	}

	return &BaseService{
		Name:      name,
		Tracer:    tp.Tracer(name),
		Provider:  tp,
		Generator: data.NewGenerator(),
		Logger:    cfg.Logger.With("service", name),
		cfg:       cfg,
	}, nil
}

func (s *BaseService) Simulate(ctx context.Context) {
	for {
		// Create a new context for each request to ensure a new trace ID
		requestCtx := context.Background()
		s.simulateChirperRequest(requestCtx)

		select {
		case <-ctx.Done():
			s.Logger.Info("Simulation stopped")
			return
		default:
			time.Sleep(s.Generator.GenerateLatency(100, 1000))
		}
	}
}

func (s *BaseService) simulateChirperRequest(ctx context.Context) {
	clientType := "web"
	if s.Generator.GenerateFloat32() < 0.3 {
		clientType = "mobile"
	}

	deviceInfo := s.generateDeviceInfo(clientType)
	userID := s.Generator.GenerateUserID()

	method := s.generateHTTPMethod()
	path := s.generateHTTPPath()

	spanName := fmt.Sprintf("%s %s", method, path)

	clientCtx, clientSpan := s.Tracer.Start(ctx, spanName,
		trace.WithAttributes(
			semconv.ServiceNameKey.String("chirper.client"),
			semconv.HTTPRequestMethodKey.String(method),
			semconv.ServerAddressKey.String("api.chirper.com"),
			semconv.ServerPortKey.Int(443),
			semconv.URLFullKey.String(fmt.Sprintf("https://api.chirper.com%s", path)),
			semconv.NetworkProtocolNameKey.String("http"),
			semconv.NetworkProtocolVersionKey.String("2.0"),
			attribute.String("client.type", clientType),
			attribute.String("user.id", userID),
		))
	defer clientSpan.End()

	for k, v := range deviceInfo {
		clientSpan.SetAttributes(attribute.String(k, v))
	}

	time.Sleep(s.Generator.GenerateLatency(10, 50))

	clientSpan.AddEvent("request_started", trace.WithAttributes(
		attribute.String("user_id", userID),
		attribute.String("request_path", path),
	))

	gatewayCtx := s.simulateAPIGateway(clientCtx, userID, method, path)

	clientSpan.AddEvent("request_completed", trace.WithAttributes(
		attribute.String("user_id", userID),
		attribute.String("request_path", path),
	))

	clientSpan.AddLink(trace.LinkFromContext(gatewayCtx))

	statusCode := s.Generator.GenerateStatusCode()
	clientSpan.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))

	if statusCode >= 400 {
		clientSpan.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
	} else {
		clientSpan.SetStatus(codes.Ok, "")
	}
}

func (s *BaseService) simulateAPIGateway(ctx context.Context, userID, method, path string) context.Context {
	gatewayTP, err := telemetry.NewProvider("chirper.api.gateway", s.cfg.Endpoint, s.cfg.Secure, s.cfg.Protocol, s.cfg.Headers)
	if err != nil {
		s.Logger.Error("Failed to create trace provider for API Gateway", "error", err)
		return ctx
	}
	defer gatewayTP.Shutdown(context.Background())

	gatewayTracer := gatewayTP.Tracer("chirper.api.gateway")

	gatewayCtx, gatewaySpan := gatewayTracer.Start(ctx, "handle_request",
		trace.WithAttributes(
			semconv.ServiceNameKey.String("chirper.api.gateway"),
			semconv.HTTPRequestMethodKey.String(method),
			semconv.URLPathKey.String(path),
			semconv.HTTPRouteKey.String(path),
			attribute.String("user.id", userID),
		))
	defer gatewaySpan.End()

	action := s.determineActionFromPath(method, path)
	gatewaySpan.SetAttributes(attribute.String("action", action))

	gatewaySpan.AddEvent("gateway_request_received", trace.WithAttributes(
		attribute.String("user_id", userID),
		attribute.String("request_path", path),
	))

	backendService := s.determineBackendService(action)
	backendCtx := s.nestedServiceCall(gatewayCtx, "chirper.api.gateway", backendService, action, userID, 0)

	gatewaySpan.AddLink(trace.LinkFromContext(backendCtx))

	gatewaySpan.AddEvent("gateway_request_completed", trace.WithAttributes(
		attribute.String("user_id", userID),
		attribute.String("request_path", path),
	))

	latency := s.Generator.GenerateLatency(10, 50)
	time.Sleep(latency)

	statusCode := s.Generator.GenerateStatusCode()
	gatewaySpan.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))

	if statusCode >= 400 {
		gatewaySpan.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
	} else {
		gatewaySpan.SetStatus(codes.Ok, "")
	}

	return gatewayCtx
}

func (s *BaseService) nestedServiceCall(ctx context.Context, parentServiceName, serviceName, action, userID string, depth int) context.Context {
	if depth > 5 {
		return ctx // Prevent infinite recursion
	}

	propagator := otel.GetTextMapPropagator()
	carrier := propagation.MapCarrier{}
	propagator.Inject(ctx, carrier)

	serviceCtx := propagator.Extract(context.Background(), carrier)

	serviceTP, err := telemetry.NewProvider(serviceName, s.cfg.Endpoint, s.cfg.Secure, s.cfg.Protocol, s.cfg.Headers)
	if err != nil {
		s.Logger.Error("Failed to create trace provider for service", "service", serviceName, "error", err)
		return ctx
	}
	defer serviceTP.Shutdown(context.Background())

	serviceTracer := serviceTP.Tracer(serviceName)

	serviceCtx, serviceSpan := serviceTracer.Start(serviceCtx, serviceName+"_operation",
		trace.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.PeerServiceKey.String(parentServiceName),
			attribute.String("action", action),
			attribute.String("user.id", userID),
		))
	defer serviceSpan.End()

	s.Logger.Info("Service call", "depth", depth, "from", parentServiceName, "to", serviceName)

	serviceSpan.AddEvent("service_operation_started", trace.WithAttributes(
		attribute.String("action", action),
		attribute.String("user_id", userID),
	))

	s.simulateBackendOperations(serviceCtx, serviceTracer, action)

	if s.Generator.GenerateFloat32() < 0.3 && depth < 3 {
		nextService := s.cfg.ServiceList[s.Generator.GenerateInt(len(s.cfg.ServiceList))]
		nextServiceCtx := s.nestedServiceCall(serviceCtx, serviceName, nextService, action, userID, depth+1)
		serviceSpan.AddLink(trace.LinkFromContext(nextServiceCtx))
	}

	if s.Generator.GenerateFloat32() < 0.2 {
		eventCtx := s.simulateCloudEvent(serviceCtx, action, userID)
		serviceSpan.AddLink(trace.LinkFromContext(eventCtx))
	}

	serviceSpan.AddEvent("service_operation_completed", trace.WithAttributes(
		attribute.String("action", action),
		attribute.String("user_id", userID),
	))

	latency := s.Generator.GenerateLatency(20, 100)
	time.Sleep(latency)

	if s.Generator.GenerateFloat32() < 0.05 {
		serviceSpan.SetStatus(codes.Error, "Service error occurred")
	} else {
		serviceSpan.SetStatus(codes.Ok, "")
	}

	return serviceCtx
}

func (s *BaseService) simulateBackendOperations(ctx context.Context, tracer trace.Tracer, action string) {
	s.simulateDatabaseOperation(ctx, tracer, action)
	s.simulateCacheOperation(ctx, tracer, action)
	s.simulateEventPublication(ctx, tracer, action)
}

func (s *BaseService) simulateCloudEvent(ctx context.Context, action, userID string) context.Context {
	eventTP, err := telemetry.NewProvider("chirper.event.service", s.cfg.Endpoint, s.cfg.Secure, s.cfg.Protocol, s.cfg.Headers)
	if err != nil {
		s.Logger.Error("Failed to create trace provider for event service", "error", err)
		return ctx
	}
	defer eventTP.Shutdown(context.Background())

	eventTracer := eventTP.Tracer("chirper.event.service")

	eventCtx, eventSpan := eventTracer.Start(ctx, "process_cloud_event",
		trace.WithAttributes(
			semconv.ServiceNameKey.String("chirper.event.service"),
			attribute.String("event.type", s.getEventType(action)),
			attribute.String("user.id", userID),
		))
	defer eventSpan.End()

	eventSpan.AddEvent("cloud_event_received", trace.WithAttributes(
		attribute.String("event.type", s.getEventType(action)),
		attribute.String("user.id", userID),
	))

	time.Sleep(s.Generator.GenerateLatency(5, 20))

	eventSpan.AddEvent("cloud_event_processed", trace.WithAttributes(
		attribute.String("event.type", s.getEventType(action)),
		attribute.String("user.id", userID),
	))

	return eventCtx
}

func (s *BaseService) simulateDatabaseOperation(ctx context.Context, tracer trace.Tracer, action string) {
	ctx, span := tracer.Start(ctx, "database_operation",
		trace.WithAttributes(
			semconv.DBSystemMySQL,
			semconv.DBNamespace("chirper_db"),
			semconv.DBOperationName(s.getDatabaseOperation(action)),
		))
	defer span.End()

	latency := s.Generator.GenerateLatency(5, 30)
	time.Sleep(latency)

	if s.Generator.GenerateFloat32() < 0.02 {
		span.SetStatus(codes.Error, "Database error")
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func (s *BaseService) simulateCacheOperation(ctx context.Context, tracer trace.Tracer, action string) {
	ctx, span := tracer.Start(ctx, "cache_operation",
		trace.WithAttributes(
			attribute.String("cache.system", "redis"),
			attribute.String("cache.operation", s.getCacheOperation(action)),
		))
	defer span.End()

	latency := s.Generator.GenerateLatency(1, 10)
	time.Sleep(latency)

	if s.Generator.GenerateFloat32() < 0.01 {
		span.SetStatus(codes.Error, "Cache error")
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func (s *BaseService) simulateEventPublication(ctx context.Context, tracer trace.Tracer, action string) {
	ctx, span := tracer.Start(ctx, "event_publication",
		trace.WithAttributes(
			attribute.String("event.system", "kafka"),
			attribute.String("event.type", s.getEventType(action)),
		))
	defer span.End()

	latency := s.Generator.GenerateLatency(1, 5)
	time.Sleep(latency)

	if s.Generator.GenerateFloat32() < 0.01 {
		span.SetStatus(codes.Error, "Event publication error")
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

func (s *BaseService) determineActionFromPath(method, path string) string {
	switch {
	case path == "/timeline":
		return "view_timeline"
	case method == "POST" && path == "/chirp":
		return "post_chirp"
	case path == "/chirp/{chirpId}/like":
		return "like_chirp"
	case path == "/chirp/{chirpId}/retweet":
		return "retweet"
	case path == "/user/{userId}/follow":
		return "follow_user"
	case path == "/user/{userId}/profile" && method == "PUT":
		return "update_profile"
	default:
		return "unknown_action"
	}
}

func (s *BaseService) determineBackendService(action string) string {
	switch action {
	case "view_timeline":
		return "chirper.timeline.service"
	case "post_chirp":
		return "chirper.post.service"
	case "like_chirp", "retweet":
		return "chirper.interaction.service"
	case "follow_user":
		return "chirper.user.service"
	case "update_profile":
		return "chirper.profile.service"
	default:
		return "chirper.general.service"
	}
}

func (s *BaseService) generateDeviceInfo(clientType string) map[string]string {
	info := make(map[string]string)
	if clientType == "web" {
		browsers := []string{"Chrome", "Firefox", "Safari"}
		oses := []string{"Windows", "MacOS", "Linux"}
		info["client.browser"] = browsers[s.Generator.GenerateInt(len(browsers))]
		info["client.os"] = oses[s.Generator.GenerateInt(len(oses))]
	} else {
		devices := []string{"iPhone", "Android"}
		info["client.device"] = devices[s.Generator.GenerateInt(len(devices))]
		info["client.os"] = info["client.device"]
	}
	return info
}

func (s *BaseService) getDatabaseOperation(action string) string {
	switch action {
	case "view_timeline":
		return "SELECT"
	case "post_chirp":
		return "INSERT"
	case "like_chirp", "retweet", "follow_user":
		return "UPDATE"
	case "update_profile":
		return "UPDATE"
	default:
		return "SELECT"
	}
}

func (s *BaseService) getCacheOperation(action string) string {
	switch action {
	case "view_timeline":
		return "GET"
	case "post_chirp", "like_chirp", "retweet":
		return "SET"
	case "follow_user", "update_profile":
		return "DEL"
	default:
		return "GET"
	}
}

func (s *BaseService) getEventType(action string) string {
	switch action {
	case "post_chirp":
		return "chirp_created"
	case "like_chirp":
		return "chirp_liked"
	case "retweet":
		return "chirp_retweeted"
	case "follow_user":
		return "user_followed"
	case "update_profile":
		return "profile_updated"
	default:
		return "general_event"
	}
}

func (s *BaseService) generateHTTPMethod() string {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	return methods[s.Generator.GenerateInt(len(methods))]
}

func (s *BaseService) generateHTTPPath() string {
	paths := []string{
		"/timeline",
		"/chirp",
		"/user/{userId}",
		"/chirp/{chirpId}/like",
		"/chirp/{chirpId}/retweet",
		"/user/{userId}/follow",
		"/user/{userId}/profile",
	}
	return paths[s.Generator.GenerateInt(len(paths))]
}

func (s *BaseService) Shutdown(ctx context.Context) error {
	s.Logger.Info("Shutting down service")
	return s.Provider.Shutdown(ctx)
}
