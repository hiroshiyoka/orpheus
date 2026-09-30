package collector

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchCloudflareMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("auth %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"viewer":{"zones":[{"httpRequestsAdaptiveGroups":[{"count":100,"sum":{"errors":5},"quantiles":{"p50":120,"p99":300}}]}]}}}`))
	}))
	defer server.Close()
	orig := cloudflareEndpoint
	cloudflareEndpoint = server.URL
	defer func() { cloudflareEndpoint = orig }()

	count, errs, p50, p99, err := FetchCloudflareMetrics("test-token", "zone123", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if count != 100 || errs != 5 || p50 == nil || *p50 != 120 || p99 == nil || *p99 != 300 {
		t.Fatalf("got %d %d %v %v", count, errs, p50, p99)
	}
}

func TestFetchCloudflareMetricsMissingToken(t *testing.T) {
	_, _, _, _, err := FetchCloudflareMetrics("", "zone", time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchCloudflareMetricsNoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"viewer":{"zones":[]}}}`))
	}))
	defer server.Close()
	orig := cloudflareEndpoint
	cloudflareEndpoint = server.URL
	defer func() { cloudflareEndpoint = orig }()
	count, errs, p50, p99, err := FetchCloudflareMetrics("token", "zone", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 || errs != 0 || p50 != nil || p99 != nil {
		t.Fatalf("expected empty")
	}
}
