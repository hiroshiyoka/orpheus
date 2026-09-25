package api

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/storage"
)

func handleProjects(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/api/projects" {
			http.NotFound(w, r)
			return
		}
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
	}
}
