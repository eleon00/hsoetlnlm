package api

import (
	"net/http"
	"strings"
	// Keep validator import if needed elsewhere, or remove if only used in handlers
	// "github.com/go-playground/validator/v10"
)

// NewRouter creates and configures a new HTTP router using the provided APIHandler.
// The handler *APIHandler parameter now correctly refers to the struct defined in handlers.go
func NewRouter(handler *APIHandler) *http.ServeMux {
	router := http.NewServeMux()

	// Health check endpoint
	router.HandleFunc("/healthz", handler.HealthCheckHandler)

	// Connections endpoints
	// Need to handle different methods and paths (/connections vs /connections/{id})
	router.HandleFunc("/connections", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.ListConnectionsHandler(w, r)
		case http.MethodPost:
			handler.CreateConnectionHandler(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	})

	router.HandleFunc("/connections/", func(w http.ResponseWriter, r *http.Request) {
		// Basic check for paths like /connections/{id}
		if !strings.Contains(r.URL.Path, "/connections/") || r.URL.Path == "/connections/" {
			http.NotFound(w, r)
			return
		}
		// Note: This simple routing doesn't validate the {id} part is numeric here.
		// It relies on the handler to parse and validate the ID.
		switch r.Method {
		case http.MethodGet:
			handler.GetConnectionHandler(w, r)
		case http.MethodPut:
			handler.UpdateConnectionHandler(w, r)
		case http.MethodDelete:
			handler.DeleteConnectionHandler(w, r)
		default:
			w.Header().Set("Allow", "GET, PUT, DELETE")
			respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	})

	// Replication Tasks endpoints
	router.HandleFunc("/replication-tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.ListReplicationTasksHandler(w, r)
		case http.MethodPost:
			handler.CreateReplicationTaskHandler(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	})

	router.HandleFunc("/replication-tasks/", func(w http.ResponseWriter, r *http.Request) {
		// Handle paths like /replication-tasks/{id} AND /replication-tasks/{task_id}/runs
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(pathParts) == 2 && pathParts[0] == "replication-tasks" {
			// Assumed /replication-tasks/{id} - Delegate to existing handler
			// Note: Simple routing, relies on handler to validate ID.
			switch r.Method {
			case http.MethodGet:
				handler.GetReplicationTaskHandler(w, r)
			case http.MethodPut:
				handler.UpdateReplicationTaskHandler(w, r)
			case http.MethodDelete:
				handler.DeleteReplicationTaskHandler(w, r)
			default:
				w.Header().Set("Allow", "GET, PUT, DELETE")
				respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			}
		} else if len(pathParts) == 3 && pathParts[0] == "replication-tasks" && pathParts[2] == "runs" {
			// Assumed /replication-tasks/{task_id}/runs
			if r.Method == http.MethodGet {
				handler.ListReplicationRunsHandler(w, r)
			} else {
				w.Header().Set("Allow", "GET")
				respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
			}
		} else {
			http.NotFound(w, r)
		}
	})

	// Benthos Configs endpoints
	router.HandleFunc("/benthos-configs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.ListBenthosConfigsHandler(w, r)
		case http.MethodPost:
			handler.CreateBenthosConfigHandler(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	})

	router.HandleFunc("/benthos-configs/", func(w http.ResponseWriter, r *http.Request) {
		pathPrefix := "/benthos-configs/"
		if !strings.HasPrefix(r.URL.Path, pathPrefix) || r.URL.Path == pathPrefix {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handler.GetBenthosConfigHandler(w, r)
		case http.MethodPut:
			handler.UpdateBenthosConfigHandler(w, r)
		case http.MethodDelete:
			handler.DeleteBenthosConfigHandler(w, r)
		default:
			w.Header().Set("Allow", "GET, PUT, DELETE")
			respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	})

	// Replication Runs endpoints
	router.HandleFunc("/replication-runs/", func(w http.ResponseWriter, r *http.Request) {
		// Route for GET /replication-runs/{run_id}
		if r.Method == http.MethodGet {
			// Basic check for path structure
			pathPrefix := "/replication-runs/"
			if !strings.HasPrefix(r.URL.Path, pathPrefix) || r.URL.Path == pathPrefix {
				http.NotFound(w, r)
				return
			}
			handler.GetReplicationRunHandler(w, r)
		} else {
			w.Header().Set("Allow", "GET")
			respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	})

	// Placeholder for other resource routes

	return router
}
