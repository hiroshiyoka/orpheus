package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/projects", handleProjects(db))
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	return mux
}
