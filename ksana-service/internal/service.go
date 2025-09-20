package internal

import (
	"context"
	"fmt"
	"ksana-service/internal/api"
	"ksana-service/internal/executor"
	"ksana-service/internal/scheduler"
	"ksana-service/internal/store"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Service struct {
	server    *http.Server
	scheduler *scheduler.Scheduler
	executor  *executor.HTTPExecutor
	store     store.Store
	logger    *slog.Logger
}

type Config struct {
	Port           string
	DataDir        string
	Workers        int
	DefaultTimeout time.Duration
	MaxRetries     int
	RetryBackoff   time.Duration
	LogLevel       string
}

func NewService(config Config) (*Service, error) {
	logLevel := slog.LevelInfo
	switch config.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	store := store.NewJSONStore(config.DataDir)

	executor := executor.NewHTTPExecutor(
		config.Workers,
		config.DefaultTimeout,
		store,
		logger,
	)

	clock := &scheduler.RealClock{}
	schedulerSvc := scheduler.NewScheduler(store, executor, clock, logger)

	handler := api.NewJobHandler(store, schedulerSvc, logger)
	router := api.NewRouter(handler, logger)

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}

	return &Service{
		server:    server,
		scheduler: schedulerSvc,
		executor:  executor,
		store:     store,
		logger:    logger,
	}, nil
}

func (s *Service) Start() error {
	s.logger.Info("Starting Ksana scheduler service", "port", s.server.Addr)

	if _, err := s.store.Load(context.Background()); err != nil {
		return fmt.Errorf("failed to load store: %w", err)
	}

	if err := s.scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", "error", err)
		}
	}()

	return nil
}

func (s *Service) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	s.logger.Info("Received shutdown signal, starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Info("Stopping HTTP server...")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown HTTP server", "error", err)
	}

	s.logger.Info("Stopping scheduler...")
	s.scheduler.Stop()

	s.logger.Info("Waiting for executor to finish...")
	if err := s.executor.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown executor", "error", err)
	}

	s.logger.Info("Graceful shutdown completed")
}

func LoadConfigFromEnv() Config {
	config := Config{
		Port:           getEnv("PORT", "7100"),
		DataDir:        getEnv("DATA_DIR", getEnv("KSANA_DATA", "./data")),
		Workers:        getEnvInt("WORKERS", 4),
		DefaultTimeout: getEnvDuration("DEFAULT_TIMEOUT", 10*time.Second),
		MaxRetries:     getEnvInt("MAX_RETRIES", 3),
		RetryBackoff:   getEnvDuration("RETRY_BACKOFF", 5*time.Second),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}