package collector

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/alerting"
	"github.com/hiroshiyoka/orpheus/internal/detector"
	"github.com/hiroshiyoka/orpheus/internal/storage"
)

var cloudflareEndpoint = "https://api.cloudflare.com/client/v4/graphql"

func FetchCloudflareMetrics(token, zoneID string, since time.Time) (int, int, *int, *int, error) {
	if token == "" {
		return 0, 0, nil, nil, fmt.Errorf("missing cloudflare token")
	}
	query := map[string]string{
		"query": fmt.Sprintf(`query{viewer{zones(filter:{zoneTag:"%s"}){httpRequestsAdaptiveGroups(limit:1, filter:{datetime_geq:"%s"}){count sum{errors} quantiles{p50 p99}}}}}}`, zoneID, since.Format(time.RFC3339)),
	}
	body, err := json.Marshal(query)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPost, cloudflareEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return 0, 0, nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return 0, 0, nil, nil, fmt.Errorf("cloudflare API returned %s", resp.Status)
	}
	var result struct {
		Data struct {
			Viewer struct {
				Zones []struct {
					Groups []struct {
						Count int `json:"count"`
						Sum   struct {
							Errors int `json:"errors"`
						} `json:"sum"`
						Quantiles struct {
							P50 float64 `json:"p50"`
							P99 float64 `json:"p99"`
						} `json:"quantiles"`
					} `json:"httpRequestsAdaptiveGroups"`
				} `json:"zones"`
			} `json:"viewer"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, nil, nil, err
	}
	if len(result.Errors) > 0 {
		return 0, 0, nil, nil, fmt.Errorf("cloudflare error: %s", result.Errors[0].Message)
	}
	if len(result.Data.Viewer.Zones) == 0 || len(result.Data.Viewer.Zones[0].Groups) == 0 {
		return 0, 0, nil, nil, nil
	}
	g := result.Data.Viewer.Zones[0].Groups[0]
	var p50, p99 *int
	if g.Quantiles.P50 != 0 {
		v := int(g.Quantiles.P50)
		p50 = &v
	}
	if g.Quantiles.P99 != 0 {
		v := int(g.Quantiles.P99)
		p99 = &v
	}
	return g.Count, g.Sum.Errors, p50, p99, nil
}

func FetchMetricsForProjects(db *sql.DB, token string, since time.Time) ([]storage.Metric, error) {
	projects, err := storage.ListProjects(db)
	if err != nil {
		return nil, err
	}
	var metrics []storage.Metric
	for _, p := range projects {
		if p.CloudflareZoneID == nil || *p.CloudflareZoneID == "" {
			continue
		}
		count, errs, p50, p99, err := FetchCloudflareMetrics(token, *p.CloudflareZoneID, since)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, storage.Metric{
			ProjectID:     p.ID,
			PeriodStart:   since,
			RequestsCount: count,
			ErrorCount:    errs,
			P50ResponseMS: p50,
			P99ResponseMS: p99,
		})
	}
	return metrics, nil
}

func StartCloudflareCollector(ctx context.Context, db *sql.DB, token string, interval time.Duration, botToken, chatID string, sendAlert AlertSender) {
	if token == "" {
		return
	}
	if sendAlert == nil {
		sendAlert = alerting.SendTelegramAlert
	}
	if interval <= 0 {
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	collect := func() {
		since := time.Now().UTC().Add(-interval)
		metrics, err := FetchMetricsForProjects(db, token, since)
		if err != nil {
			log.Printf("cloudflare fetch failed: %v", err)
			return
		}
		for _, m := range metrics {
			if _, err := storage.InsertMetric(db, m); err != nil {
				log.Printf("metric insert failed: %v", err)
				continue
			}
			project, err := storage.GetProject(db, m.ProjectID)
			if err != nil {
				continue
			}
			if m.RequestsCount > 0 && float64(m.ErrorCount)/float64(m.RequestsCount) > 0.05 {
				incident, err := detector.CheckErrorSpike(db, m.ProjectID, m.ErrorCount, m.RequestsCount, 0.05)
				if err != nil {
					log.Printf("error spike check failed: %v", err)
				} else if incident != nil {
					message := fmt.Sprintf("[%s] Error spike\nURL: %s", project.Name, project.URL)
					sent, err := storage.AlertSent(db, incident.ID, "telegram", message)
					if err != nil {
						log.Printf("alert lookup failed: %v", err)
					} else if !sent {
						if err := sendAlert(botToken, chatID, message); err != nil {
							log.Printf("telegram alert failed: %v", err)
						} else if err := storage.InsertAlert(db, incident.ID, "telegram", message); err != nil {
							log.Printf("alert log failed: %v", err)
						}
					}
				}
			} else {
				incident, err := detector.ResolveErrorSpike(db, m.ProjectID, m.ErrorCount, m.RequestsCount, 0.05)
				if err != nil {
					log.Printf("error spike resolve failed: %v", err)
				} else if incident != nil {
					message := fmt.Sprintf("[%s] Error recovered\nURL: %s", project.Name, project.URL)
					sent, err := storage.AlertSent(db, incident.ID, "telegram", message)
					if err != nil {
						log.Printf("alert lookup failed: %v", err)
					} else if !sent {
						if err := sendAlert(botToken, chatID, message); err != nil {
							log.Printf("telegram alert failed: %v", err)
						} else if err := storage.InsertAlert(db, incident.ID, "telegram", message); err != nil {
							log.Printf("alert log failed: %v", err)
						}
					}
				}
			}
		}
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			collect()
		}
	}
}
