package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eleon00/hsoetlnlm/internal/api"
	"github.com/eleon00/hsoetlnlm/internal/data"
	"github.com/eleon00/hsoetlnlm/internal/service"
	"github.com/eleon00/hsoetlnlm/internal/temporal"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize structured logger (zerolog)
	// Configure for console output with reasonable defaults
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	logger.Info().Msg("Application starting up...")

	// Configuration from Environment Variables
	dsn := os.Getenv("APP_DB_DSN")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/hsoetlnlm_db?sslmode=disable" // Default for local non-docker
		logger.Warn().Str("dsn", dsn).Msg("APP_DB_DSN not set, using default")
	}

	temporalHostPort := os.Getenv("TEMPORAL_HOST_PORT")
	if temporalHostPort == "" {
		temporalHostPort = "localhost:7233" // Default for local non-docker
		logger.Warn().Str("hostPort", temporalHostPort).Msg("TEMPORAL_HOST_PORT not set, using default")
	}

	// Initialize Data Layer
	// Note: data.Repository is an interface. We assume *data.DB implements it.
	// If NewDB returns an error, it might be because the placeholder DSN is used
	// or the database is not reachable/configured correctly.
	db, err := data.NewDB(dsn)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to initialize database connection. Continuing without DB...")
		// Handle the case where DB connection fails. Maybe run in a limited mode?
		// For now, we'll proceed, but repository operations will fail.
		db = nil // Ensure db is nil if connection failed
	} else {
		// Only defer close if db was successfully initialized
		defer func() {
			if err := db.Close(); err != nil {
				logger.Error().Err(err).Msg("Error closing database")
			}
		}()
		logger.Info().Msg("Database connection pool initialized (or prepared).")
	}

	// Initialize Service Layer
	// Pass the db instance, which should satisfy the data.Repository interface.
	// If db is nil due to connection failure, the service layer might
	// need to handle this gracefully or fail operations that require the DB.
	appService := service.NewService(db) // db might be nil here!
	logger.Info().Msg("Service layer initialized.")

	// Initialize Temporal client (optional)
	var temporalClient *temporal.Client
	// Uncomment to enable Temporal
	// Make sure TEMPORAL_HOST_PORT env var is set correctly for docker-compose (e.g., "temporal:7233")
	// or leave unset for local default ("localhost:7233")
	/*
		temporalClient, err = temporal.NewClient(&temporal.ClientOptions{
			HostPort:    temporalHostPort, // Use host/port from env/default
			Namespace:   "default",
			ServiceName: "hsoetlnlm",
		})
		if err != nil {
			logger.Warn().Err(err).Str("hostPort", temporalHostPort).Msg("Failed to initialize Temporal client. Continuing without Temporal...")
		} else {
			defer temporalClient.Close()
			logger.Info().Str("hostPort", temporalHostPort).Msg("Temporal client initialized.")
			service.WorkflowClientImpl = temporalClient
		}
	*/

	// Initialize Temporal worker (optional)
	var temporalWorker *temporal.Worker
	if temporalClient != nil {
		temporalWorker, err = temporal.NewWorker(temporalClient, appService, &temporal.WorkerOptions{
			TaskQueue: "replication-tasks",
		})
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to initialize Temporal worker")
		} else {
			// Start the worker
			err = temporalWorker.Start()
			if err != nil {
				logger.Warn().Err(err).Msg("Failed to start Temporal worker")
			} else {
				logger.Info().Msg("Temporal worker started successfully.")
				defer temporalWorker.Stop()
			}
		}
	}

	// Initialize API Layer - PASS THE LOGGER HERE
	apiHandler := api.NewAPIHandler(appService, logger)
	logger.Info().Msg("API handler initialized.")

	// Create the main router
	router := api.NewRouter(apiHandler)
	logger.Info().Msg("Router initialized.")

	// Define the port the server will listen on
	port := ":8080"
	logger.Info().Str("port", port).Msgf("Starting server on http://localhost%s", port)

	// Create HTTP server
	server := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so it doesn't block
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()
	logger.Info().Msg("Server started successfully.")

	// Setup graceful shutdown
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("Server is shutting down...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server shutdown failed")
	}
	logger.Info().Msg("Server shutdown complete")
}
