package collector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/storage"
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

func TestFetchMetricsForProjects(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	zone := "zone-abc"
	if _, err := storage.CreateProject(db, storage.Project{Name: "with-zone", URL: "https://example.com", CloudflareZoneID: &zone, IsActive: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.CreateProject(db, storage.Project{Name: "without-zone", URL: "https://example.org", IsActive: true}); err != nil {
		t.Fatal(err)
	}
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Write([]byte(`{"data":{"viewer":{"zones":[{"httpRequestsAdaptiveGroups":[{"count":50,"sum":{"errors":2},"quantiles":{"p50":80,"p99":200}}]}]}}}`))
	}))
	defer server.Close()
	orig := cloudflareEndpoint
	cloudflareEndpoint = server.URL
	defer func() { cloudflareEndpoint = orig }()
	metrics, err := FetchMetricsForProjects(db, "token", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].RequestsCount != 50 || metrics[0].ErrorCount != 2 {
		t.Fatalf("unexpected metric %+v", metrics[0])
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestStartCloudflareCollector(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	zone := "zone-xyz"
	p, err := storage.CreateProject(db, storage.Project{Name: "cf-project", URL: "https://example.com", CloudflareZoneID: &zone, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"viewer":{"zones":[{"httpRequestsAdaptiveGroups":[{"count":10,"sum":{"errors":1},"quantiles":{"p50":90,"p99":210}}]}]}}}`))
	}))
	defer server.Close()
	orig := cloudflareEndpoint
	cloudflareEndpoint = server.URL
	defer func() { cloudflareEndpoint = orig }()
	ctx, cancel := context.WithCancel(context.Background())
	go StartCloudflareCollector(ctx, db, "token", 100*time.Millisecond)
	time.Sleep(350 * time.Millisecond)
	cancel()
	metrics, err := storage.GetMetricsByProject(db, p.ID, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) == 0 {
		t.Fatal("expected metrics")
	}
}
