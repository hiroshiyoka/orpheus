package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestMetricStorage(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	project, err := CreateProject(db, Project{Name: "Example", URL: "https://example.com", CheckIntervalSeconds: 60, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p50, p99 := 120, 500
	first, err := InsertMetric(db, Metric{
		ProjectID:     project.ID,
		PeriodStart:   start,
		RequestsCount: 1000,
		ErrorCount:    5,
		P50ResponseMS: &p50,
		P99ResponseMS: &p99,
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == 0 || first.RequestsCount != 1000 || first.P50ResponseMS == nil || *first.P50ResponseMS != p50 {
		t.Fatalf("unexpected inserted metric: %+v", first)
	}

	if _, err := InsertMetric(db, Metric{ProjectID: project.ID, PeriodStart: start.Add(time.Hour), RequestsCount: 800, ErrorCount: 2}); err != nil {
		t.Fatal(err)
	}
	metrics, err := GetMetricsByProject(db, project.ID, start.Add(30*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 1 || metrics[0].RequestsCount != 800 || metrics[0].P50ResponseMS != nil {
		t.Fatalf("unexpected filtered metrics: %+v", metrics)
	}
}
