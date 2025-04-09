package api

import (
	"database/sql" // Import necessary for checking sql.ErrNoRows
	"encoding/json"
	"errors" // Import necessary for errors.Is
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/eleon00/hsoetlnlm/internal/data"
	"github.com/eleon00/hsoetlnlm/internal/service"
	"github.com/go-playground/validator/v10" // Import validator
	"github.com/rs/zerolog"                  // Import zerolog
)

// APIHandler holds dependencies for API handlers, like the service layer.
type APIHandler struct {
	svc       service.Service
	validator *validator.Validate // Add validator field
	logger    zerolog.Logger      // Add logger field
}

// NewAPIHandler creates a new APIHandler with the necessary dependencies.
func NewAPIHandler(svc service.Service, logger zerolog.Logger) *APIHandler {
	return &APIHandler{
		svc:       svc,
		validator: validator.New(), // Initialize validator
		logger:    logger,          // Store the logger
	}
}

// --- Helper Functions ---

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal Server Error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// --- Handlers ---

// HealthCheckHandler handles requests to the /healthz endpoint.
func (h *APIHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// In a real scenario, this might call h.svc.HealthCheck() or similar
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListConnectionsHandler handles GET requests to /connections.
func (h *APIHandler) ListConnectionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	connections, err := h.svc.ListConnections(r.Context())
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error listing connections")
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve connections")
		return
	}

	respondWithJSON(w, http.StatusOK, connections)
}

// CreateConnectionHandler handles POST requests to /connections.
func (h *APIHandler) CreateConnectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var input data.Connection
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error decoding create connection request")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// --- Input Validation ---
	err = h.validator.Struct(input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Use injected logger
			h.logger.Warn().Err(err).Interface("validation_errors", validationErrors).Msg("Validation errors for create connection")
			// TODO: Consider returning specific validation errors to client
			respondWithError(w, http.StatusBadRequest, "Invalid input: validation failed")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Msg("Unexpected error during validation")
			respondWithError(w, http.StatusInternalServerError, "Validation error")
		}
		return
	}
	// --- End Input Validation ---

	newID, err := h.svc.CreateConnection(r.Context(), &input)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error creating connection")
		respondWithError(w, http.StatusInternalServerError, "Failed to create connection")
		return
	}

	// Set Location header for the newly created resource
	w.Header().Set("Location", fmt.Sprintf("/connections/%d", newID))

	// Respond with the created object, potentially fetching it again to get all fields
	// For now, just respond with the input data + new ID
	input.ID = newID
	respondWithJSON(w, http.StatusCreated, input)
}

// GetConnectionHandler handles GET requests to /connections/{id}.
func (h *APIHandler) GetConnectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Extract ID from URL path (naive implementation, needs improvement for routing)
	idStr := r.URL.Path[len("/connections/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid connection ID")
		return
	}

	connection, err := h.svc.GetConnection(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Connection not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("connection_id", id).Msg("Error getting connection")
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve connection")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, connection)
}

// UpdateConnectionHandler handles PUT requests to /connections/{id}.
func (h *APIHandler) UpdateConnectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Extract ID from URL path
	idStr := r.URL.Path[len("/connections/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid connection ID")
		return
	}

	var input data.Connection
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error decoding update connection request")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	input.ID = id

	// --- Input Validation ---
	err = h.validator.Struct(input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Use injected logger
			h.logger.Warn().Err(err).Int64("connection_id", id).Interface("validation_errors", validationErrors).Msg("Validation errors for update connection")
			respondWithError(w, http.StatusBadRequest, "Invalid input: validation failed")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("connection_id", id).Msg("Unexpected error during validation for update connection")
			respondWithError(w, http.StatusInternalServerError, "Validation error")
		}
		return
	}
	// --- End Input Validation ---

	err = h.svc.UpdateConnection(r.Context(), &input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Assuming service/repo might return this
			respondWithError(w, http.StatusNotFound, "Connection not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("connection_id", id).Msg("Error updating connection")
			respondWithError(w, http.StatusInternalServerError, "Failed to update connection")
		}
		return
	}

	// Respond with the updated object, potentially fetching it again
	// For now, just respond with the input data
	respondWithJSON(w, http.StatusOK, input)
}

// DeleteConnectionHandler handles DELETE requests to /connections/{id}.
func (h *APIHandler) DeleteConnectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Extract ID from URL path
	idStr := r.URL.Path[len("/connections/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid connection ID")
		return
	}

	err = h.svc.DeleteConnection(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Assuming service/repo might return this
			respondWithError(w, http.StatusNotFound, "Connection not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("connection_id", id).Msg("Error deleting connection")
			respondWithError(w, http.StatusInternalServerError, "Failed to delete connection")
		}
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil) // 204 No Content on successful deletion
}

// --- Replication Task Handlers ---

// ListReplicationTasksHandler handles GET requests to /replication-tasks.
func (h *APIHandler) ListReplicationTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	tasks, err := h.svc.ListReplicationTasks(r.Context())
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error listing replication tasks")
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve replication tasks")
		return
	}

	respondWithJSON(w, http.StatusOK, tasks)
}

// CreateReplicationTaskHandler handles POST requests to /replication-tasks.
func (h *APIHandler) CreateReplicationTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var input data.ReplicationTask
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error decoding create replication task request")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// --- Input Validation ---
	err = h.validator.Struct(input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Use injected logger
			h.logger.Warn().Err(err).Interface("validation_errors", validationErrors).Msg("Validation errors for create replication task")
			respondWithError(w, http.StatusBadRequest, "Invalid input: validation failed")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Msg("Unexpected error during validation")
			respondWithError(w, http.StatusInternalServerError, "Validation error")
		}
		return
	}
	// --- End Input Validation ---

	newID, err := h.svc.CreateReplicationTask(r.Context(), &input)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error creating replication task")
		respondWithError(w, http.StatusInternalServerError, "Failed to create replication task")
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/replication-tasks/%d", newID))
	input.ID = newID // Add ID to the response object
	respondWithJSON(w, http.StatusCreated, input)
}

// GetReplicationTaskHandler handles GET requests to /replication-tasks/{id}.
func (h *APIHandler) GetReplicationTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Extract ID from URL path (using the same naive approach as connections)
	idStr := r.URL.Path[len("/replication-tasks/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid replication task ID")
		return
	}

	task, err := h.svc.GetReplicationTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Replication task not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("task_id", id).Msg("Error getting replication task")
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve replication task")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

// UpdateReplicationTaskHandler handles PUT requests to /replication-tasks/{id}.
func (h *APIHandler) UpdateReplicationTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Extract ID from URL path
	idStr := r.URL.Path[len("/replication-tasks/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid replication task ID")
		return
	}

	var input data.ReplicationTask
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error decoding update replication task request")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	input.ID = id // Set ID from URL

	// --- Input Validation ---
	err = h.validator.Struct(input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Use injected logger
			h.logger.Warn().Err(err).Int64("task_id", id).Interface("validation_errors", validationErrors).Msg("Validation errors for update replication task")
			respondWithError(w, http.StatusBadRequest, "Invalid input: validation failed")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("task_id", id).Msg("Unexpected error during validation for update replication task")
			respondWithError(w, http.StatusInternalServerError, "Validation error")
		}
		return
	}
	// --- End Input Validation ---

	err = h.svc.UpdateReplicationTask(r.Context(), &input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Assuming service/repo might return this
			respondWithError(w, http.StatusNotFound, "Replication task not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("task_id", id).Msg("Error updating replication task")
			respondWithError(w, http.StatusInternalServerError, "Failed to update replication task")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, input)
}

// DeleteReplicationTaskHandler handles DELETE requests to /replication-tasks/{id}.
func (h *APIHandler) DeleteReplicationTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Extract ID from URL path
	idStr := r.URL.Path[len("/replication-tasks/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid replication task ID")
		return
	}

	err = h.svc.DeleteReplicationTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Assuming service/repo might return this
			respondWithError(w, http.StatusNotFound, "Replication task not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("task_id", id).Msg("Error deleting replication task")
			respondWithError(w, http.StatusInternalServerError, "Failed to delete replication task")
		}
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

// --- Benthos Config Handlers ---

// ListBenthosConfigsHandler handles GET requests to /benthos-configs.
func (h *APIHandler) ListBenthosConfigsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	configs, err := h.svc.ListBenthosConfigs(r.Context())
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error listing benthos configs")
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve benthos configs")
		return
	}

	respondWithJSON(w, http.StatusOK, configs)
}

// CreateBenthosConfigHandler handles POST requests to /benthos-configs.
func (h *APIHandler) CreateBenthosConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var input data.BenthosConfiguration
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error decoding create benthos config request")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// --- Input Validation ---
	err = h.validator.Struct(input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Use injected logger
			h.logger.Warn().Err(err).Interface("validation_errors", validationErrors).Msg("Validation errors for create benthos config")
			respondWithError(w, http.StatusBadRequest, "Invalid input: validation failed")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Msg("Unexpected error during validation")
			respondWithError(w, http.StatusInternalServerError, "Validation error")
		}
		return
	}
	// --- End Input Validation ---

	newID, err := h.svc.CreateBenthosConfig(r.Context(), &input)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error creating benthos config")
		respondWithError(w, http.StatusInternalServerError, "Failed to create benthos config")
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/benthos-configs/%d", newID))
	input.ID = newID
	respondWithJSON(w, http.StatusCreated, input)
}

// GetBenthosConfigHandler handles GET requests to /benthos-configs/{id}.
func (h *APIHandler) GetBenthosConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	idStr := r.URL.Path[len("/benthos-configs/"):] // Naive routing
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid benthos config ID")
		return
	}

	config, err := h.svc.GetBenthosConfig(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Benthos config not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("config_id", id).Msg("Error getting benthos config")
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve benthos config")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, config)
}

// UpdateBenthosConfigHandler handles PUT requests to /benthos-configs/{id}.
func (h *APIHandler) UpdateBenthosConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	idStr := r.URL.Path[len("/benthos-configs/"):] // Naive routing
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid benthos config ID")
		return
	}

	var input data.BenthosConfiguration
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Msg("Error decoding update benthos config request")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	input.ID = id // Set ID from URL

	// --- Input Validation ---
	err = h.validator.Struct(input)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Use injected logger
			h.logger.Warn().Err(err).Int64("config_id", id).Interface("validation_errors", validationErrors).Msg("Validation errors for update benthos config")
			respondWithError(w, http.StatusBadRequest, "Invalid input: validation failed")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("config_id", id).Msg("Unexpected error during validation for update benthos config")
			respondWithError(w, http.StatusInternalServerError, "Validation error")
		}
		return
	}
	// --- End Input Validation ---

	err = h.svc.UpdateBenthosConfig(r.Context(), &input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Benthos config not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("config_id", id).Msg("Error updating benthos config")
			respondWithError(w, http.StatusInternalServerError, "Failed to update benthos config")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, input)
}

// DeleteBenthosConfigHandler handles DELETE requests to /benthos-configs/{id}.
func (h *APIHandler) DeleteBenthosConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	idStr := r.URL.Path[len("/benthos-configs/"):] // Naive routing
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid benthos config ID")
		return
	}

	err = h.svc.DeleteBenthosConfig(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Benthos config not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("config_id", id).Msg("Error deleting benthos config")
			respondWithError(w, http.StatusInternalServerError, "Failed to delete benthos config")
		}
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

// --- Replication Run Handlers ---

// ListReplicationRunsHandler handles GET requests to /replication-tasks/{task_id}/runs
func (h *APIHandler) ListReplicationRunsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Extract task ID from URL path (assuming path like /replication-tasks/{task_id}/runs)
	// This needs a proper router for robust extraction
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 || pathParts[0] != "replication-tasks" || pathParts[2] != "runs" {
		respondWithError(w, http.StatusBadRequest, "Invalid URL path format")
		return
	}
	taskIDStr := pathParts[1]
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid replication task ID")
		return
	}

	runs, err := h.svc.ListReplicationRuns(r.Context(), taskID)
	if err != nil {
		// Use injected logger
		h.logger.Error().Err(err).Int64("task_id", taskID).Msg("Error listing replication runs")
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve replication runs")
		return
	}

	respondWithJSON(w, http.StatusOK, runs)
}

// GetReplicationRunHandler handles GET requests to /replication-runs/{run_id}
func (h *APIHandler) GetReplicationRunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Extract ID from URL path
	runIDStr := r.URL.Path[len("/replication-runs/"):]
	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid replication run ID")
		return
	}

	runDetails, err := h.svc.GetReplicationRunDetails(r.Context(), runID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Replication run not found")
		} else {
			// Use injected logger
			h.logger.Error().Err(err).Int64("run_id", runID).Msg("Error getting replication run details")
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve replication run details")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, runDetails)
}
