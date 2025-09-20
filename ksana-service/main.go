package main

import (
	"ksana-service/internal"
	"log"
	"log/slog"
)

func main() {
	config := internal.LoadConfigFromEnv()

	service, err := internal.NewService(config)
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	slog.Info("Ksana scheduler service started successfully")

	service.WaitForShutdown()
}
