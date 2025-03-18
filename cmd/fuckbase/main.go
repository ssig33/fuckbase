package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ssig33/fuckbase/internal/config"
	"github.com/ssig33/fuckbase/internal/database"
	"github.com/ssig33/fuckbase/internal/logger"
	"github.com/ssig33/fuckbase/internal/server"
)

func main() {
	// Create server configuration
	cfg := config.NewServerConfig()
	cfg.Parse()

	// Initialize logger
	if err := logger.InitLogger(cfg.LogLevel, cfg.LogFile); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		logger.Error("Failed to create data directory: %v", err)
		os.Exit(1)
	}

	// Create database manager
	dbManager := database.NewManager()

	// Create HTTP server
	srv := server.NewServer(cfg, dbManager)

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("Server error: %v", err)
			stop <- os.Interrupt
		}
	}()

	logger.Info("FuckBase server started")

	// Wait for interrupt signal
	<-stop
	logger.Info("Shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Stop(ctx); err != nil {
		logger.Error("Server shutdown error: %v", err)
	}

	logger.Info("Server stopped")
}