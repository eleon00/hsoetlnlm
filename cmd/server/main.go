package main

import (
	"log"
	"net/http"

	"github.com/eleon00/hsoetlnlm/internal/api"
	"github.com/eleon00/hsoetlnlm/internal/data"
	"github.com/eleon00/hsoetlnlm/internal/service"
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

	// Initialize API Layer
	apiHandler := api.NewAPIHandler(appService)
	log.Println("API handler initialized.")

	// Create the main router
	router := api.NewRouter(apiHandler)
	log.Println("Router initialized.")

	// Define the port the server will listen on
	port := ":8080"
	log.Printf("Starting server on http://localhost%s\n", port)

	// Start the HTTP server with the router
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
