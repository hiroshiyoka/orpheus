package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
