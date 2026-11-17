package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"example.com/gin-layout/internal/app"
	"example.com/gin-layout/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Printf("API stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	api, err := app.New(cfg)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("API listening on %s", cfg.HTTPAddr)
	return api.Run(ctx)
}
