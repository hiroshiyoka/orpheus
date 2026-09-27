package api

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/storage"
)

func handleProjects(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			projects, err := storage.ListProjects(db)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			type resp struct {
				ID               int64      `json:"id"`
				Name             string     `json:"name"`
				URL              string     `json:"url"`
				IsUp             bool       `json:"is_up"`
				LastCheckedAt    *time.Time `json:"last_checked_at"`
				ResponseTimeMS   *int       `json:"response_time_ms"`
				Uptime24hPercent float64    `json:"uptime_24h_percent"`
			}
			out := make([]resp, 0, len(projects))
			for _, p := range projects {
				checks, err := storage.GetChecksByProject(db, p.ID, time.Time{})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				var isUp bool
				var last *time.Time
				var rt *int
				if len(checks) > 0 {
					latest := checks[len(checks)-1]
					isUp = latest.IsUp
					last = &latest.CheckedAt
					rt = latest.ResponseTimeMS
				}
				since := time.Now().UTC().Add(-24 * time.Hour)
				recent, err := storage.GetChecksByProject(db, p.ID, since)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				var uptime float64
				if len(recent) > 0 {
					up := 0
					for _, c := range recent {
						if c.IsUp {
							up++
						}
					}
					uptime = float64(up) / float64(len(recent)) * 100
					uptime = math.Round(uptime*10) / 10
				}
				out = append(out, resp{
					ID:               p.ID,
					Name:             p.Name,
					URL:              p.URL,
					IsUp:             isUp,
					LastCheckedAt:    last,
					ResponseTimeMS:   rt,
					Uptime24hPercent: uptime,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(out)
		case http.MethodPost:
			var req struct {
				Name                 *string `json:"name"`
				URL                  *string `json:"url"`
				CloudflareZoneID     *string `json:"cloudflare_zone_id"`
				CheckIntervalSeconds *int    `json:"check_interval_seconds"`
				FailureThreshold     *int    `json:"failure_threshold"`
				IsActive             *bool   `json:"is_active"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if req.Name == nil || *req.Name == "" || req.URL == nil || *req.URL == "" {
				http.Error(w, "name and url are required", http.StatusBadRequest)
				return
			}
			p := storage.Project{Name: *req.Name, URL: *req.URL}
			if req.CloudflareZoneID != nil {
				p.CloudflareZoneID = req.CloudflareZoneID
			}
			if req.CheckIntervalSeconds != nil {
				p.CheckIntervalSeconds = *req.CheckIntervalSeconds
			}
			if req.FailureThreshold != nil {
				p.FailureThreshold = *req.FailureThreshold
			}
			if req.IsActive != nil {
				p.IsActive = *req.IsActive
			} else {
				p.IsActive = true
			}
			created, err := storage.CreateProject(db, p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := strings.TrimPrefix(r.URL.Path, "/api/projects/")
		if idStr == "" || strings.Contains(idStr, "/") {
			http.NotFound(w, r)
			return
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		existing, err := storage.GetProject(db, id)
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var req struct {
			Name                 *string `json:"name"`
			URL                  *string `json:"url"`
			CloudflareZoneID     *string `json:"cloudflare_zone_id"`
			CheckIntervalSeconds *int    `json:"check_interval_seconds"`
			FailureThreshold     *int    `json:"failure_threshold"`
			IsActive             *bool   `json:"is_active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.Name != nil {
			existing.Name = *req.Name
		}
		if req.URL != nil {
			existing.URL = *req.URL
		}
		if req.CloudflareZoneID != nil {
			existing.CloudflareZoneID = req.CloudflareZoneID
		}
		if req.CheckIntervalSeconds != nil {
			existing.CheckIntervalSeconds = *req.CheckIntervalSeconds
		}
		if req.FailureThreshold != nil {
			existing.FailureThreshold = *req.FailureThreshold
		}
		if req.IsActive != nil {
			existing.IsActive = *req.IsActive
		}
		if err := storage.UpdateProject(db, existing); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		updated, err := storage.GetProject(db, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	}
}
