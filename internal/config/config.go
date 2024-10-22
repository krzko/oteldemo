package config

import (
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel/log"
)

type Config struct {
	Endpoint    string
	Headers     map[string]string
	Secure      bool
	Protocol    string
	ServiceList []string
	Logger      *slog.Logger
	OtelLogger  log.Logger
	Environment string
	Version     string
	Commit      string
	BuildDate   string
	ShowVersion bool
}

func Parse() *Config {
	godotenv.Load()

	cfg := &Config{
		ServiceList: make([]string, 0, 20),
		Headers:     make(map[string]string),
	}

	flag.BoolVar(&cfg.ShowVersion, "version", false, "Display version information")
	flag.BoolVar(&cfg.ShowVersion, "v", false, "Display version information (shorthand)")

	flag.StringVar(&cfg.Endpoint, "endpoint", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"), "OpenTelemetry collector endpoint")
	flag.BoolVar(&cfg.Secure, "secure", getEnvBool("OTEL_EXPORTER_OTLP_SECURE", false), "Use secure connection")
	flag.StringVar(&cfg.Protocol, "protocol", getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "http"), "Protocol to use (grpc or http)")
	flag.StringVar(&cfg.Environment, "env", getEnv("APP_ENV", "development"), "Application environment")

	headers := flag.String("headers", getEnv("OTEL_EXPORTER_OTLP_HEADERS", ""), "Headers as key=value pairs, comma-separated")

	flag.Parse()

	cfg.Headers = parseHeaders(*headers)

	services := []string{
		"client.web", "client.mobile",
		"api.gateway",
		"web.server",
		"auth.service", "user.service", "post.service", "timeline.service",
		"notification.service", "search.service", "analytics.service",
		"media.service", "trending.service", "messaging.service",
		"event.bus", "cache.service", "database.service",
	}

	for _, service := range services {
		cfg.ServiceList = append(cfg.ServiceList, "chirper."+service)
	}

	cfg.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return cfg
}

func parseHeaders(headerString string) map[string]string {
	headers := make(map[string]string)
	if headerString == "" {
		return headers
	}

	pairs := strings.Split(headerString, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			headers[key] = value
		}
	}

	return headers
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		return strings.ToLower(value) == "true"
	}
	return fallback
}
