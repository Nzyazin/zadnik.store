package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Nzyazin/zadnik.store/internal/image/app"
	"github.com/Nzyazin/zadnik.store/internal/image/config"
)

func main() {
		config, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		application, err := app.NewApp(config)
		if err != nil {
			log.Fatalf("Failed to initialize application: %v", err)
		}
		defer application.Shutdown()

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		if err := application.Run(ctx); err != nil {
			log.Fatalf("Failed to run application: %v", err)
		}
}
