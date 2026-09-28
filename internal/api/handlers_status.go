package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/storage"
)

func handleChecks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, ok := parseProjectID(w, r, "/checks")
		if !ok {
			return
		}
		if _, err := storage.GetProject(db, id); err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		since := time.Now().UTC().Add(-parseRange(r.URL.Query().Get("range"), 24*time.Hour))
		checks, err := storage.GetChecksByProject(db, id, since)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if checks == nil {
			checks = []storage.Check{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(checks)
	}
}

func handleMetrics(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id, ok := parseProjectID(w, r, "/metrics")
		if !ok {
			return
		}
		if _, err := storage.GetProject(db, id); err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		since := time.Now().UTC().Add(-parseRange(r.URL.Query().Get("range"), 7*24*time.Hour))
		metrics, err := storage.GetMetricsByProject(db, id, since)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if metrics == nil {
			metrics = []storage.Metric{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	}
}

func parseProjectID(w http.ResponseWriter, r *http.Request, suffix string) (int64, bool) {
	if !strings.HasSuffix(r.URL.Path, suffix) {
		http.NotFound(w, r)
		return 0, false
	}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	trimmed = strings.TrimSuffix(trimmed, suffix)
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" || strings.Contains(trimmed, "/") {
		http.NotFound(w, r)
		return 0, false
	}
	id, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func parseRange(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return def
		}
		return time.Duration(n) * 24 * time.Hour
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}
