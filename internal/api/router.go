package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/projects", handleProjects(db))
	mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch && !strings.Contains(strings.TrimPrefix(r.URL.Path, "/api/projects/"), "/") {
			handleProject(db).ServeHTTP(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/checks") {
			handleChecks(db).ServeHTTP(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/metrics") {
			handleMetrics(db).ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	return mux
}
