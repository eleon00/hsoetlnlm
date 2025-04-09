package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eleon00/hsoetlnlm/internal/api"
	"github.com/eleon00/hsoetlnlm/internal/data"
	"github.com/eleon00/hsoetlnlm/internal/service"
	"github.com/eleon00/hsoetlnlm/internal/temporal"
)

func main() {
	// Configuration (replace with actual config loading later)
	// For now, using a placeholder DSN. Ensure your DB is running and accessible.
	// Example DSN format for SQL Server: "sqlserver://username:password@host:port?database=dbname"
	dsn := "placeholder_dsn" // <-- IMPORTANT: Replace with your actual database connection string
	if dsn == "placeholder_dsn" {
		log.Println("WARNING: Using placeholder DSN. Update cmd/server/main.go with your actual database connection string.")
		// You might want to exit here in a real app if config is mandatory
		// log.Fatal("Database DSN is not configured.")
	}

	// Initialize Data Layer
	// Note: data.Repository is an interface. We assume *data.DB implements it.
	// If NewDB returns an error, it might be because the placeholder DSN is used
	// or the database is not reachable/configured correctly.
	db, err := data.NewDB(dsn)
	if err != nil {
		log.Printf("WARNING: Failed to initialize database connection: %v. Continuing without DB...", err)
		// Handle the case where DB connection fails. Maybe run in a limited mode?
		// For now, we'll proceed, but repository operations will fail.
		db = nil // Ensure db is nil if connection failed
	} else {
		// Only defer close if db was successfully initialized
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("Error closing database: %v", err)
			}
		}()
		log.Println("Database connection pool initialized (or prepared).")
	}

	// Initialize Service Layer
	// Pass the db instance, which should satisfy the data.Repository interface.
	// If db is nil due to connection failure, the service layer might
	// need to handle this gracefully or fail operations that require the DB.
	appService := service.NewService(db) // db might be nil here!
	log.Println("Service layer initialized.")

	// Initialize Temporal client (optional)
	var temporalClient *temporal.Client
	// Uncomment to enable Temporal (requires a running Temporal server)
	/*
		temporalClient, err = temporal.NewClient(&temporal.ClientOptions{
			HostPort:    "localhost:7233", // Default Temporal server address
			Namespace:   "default",
			ServiceName: "hsoetlnlm",
		})
		if err != nil {
			log.Printf("WARNING: Failed to initialize Temporal client: %v. Continuing without Temporal...", err)
		} else {
			defer temporalClient.Close()
			log.Println("Temporal client initialized.")

			// Set the client as the global workflow client implementation
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
			log.Printf("WARNING: Failed to initialize Temporal worker: %v", err)
		} else {
			// Start the worker
			err = temporalWorker.Start()
			if err != nil {
				log.Printf("WARNING: Failed to start Temporal worker: %v", err)
			} else {
				log.Println("Temporal worker started successfully.")
				defer temporalWorker.Stop()
			}
		}
	}

	// Initialize API Layer
	apiHandler := api.NewAPIHandler(appService)
	log.Println("API handler initialized.")

	// Create the main router
	router := api.NewRouter(apiHandler)
	log.Println("Router initialized.")

	// Define the port the server will listen on
	port := ":8080"
	log.Printf("Starting server on http://localhost%s\n", port)

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
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	log.Println("Server started successfully.")

	// Setup graceful shutdown
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server is shutting down...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server shutdown complete")
}
