package health

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type response struct {
	Status string `json:"status"`
}

// NewHealthHandler returns an http.Handler that responds to GET /healthz.
// Returns 200 {"status":"ok"} if the database is reachable, 503 otherwise.
func NewHealthHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(response{Status: "unhealthy"})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{Status: "ok"})
	})
}
