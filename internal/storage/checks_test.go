package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCheckStorage(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	project, err := CreateProject(db, Project{Name: "Example", URL: "https://example.com", CheckIntervalSeconds: 60, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	statusCode, responseTime := 200, 42
	message := "connection failed"
	firstTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	first, err := InsertCheck(db, Check{
		ProjectID:      project.ID,
		CheckedAt:      firstTime,
		IsUp:           true,
		StatusCode:     &statusCode,
		ResponseTimeMS: &responseTime,
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == 0 || !first.CheckedAt.Equal(firstTime) || first.StatusCode == nil || *first.StatusCode != 200 {
		t.Fatalf("unexpected inserted check: %+v", first)
	}

	secondTime := firstTime.Add(time.Hour)
	if _, err := InsertCheck(db, Check{ProjectID: project.ID, CheckedAt: secondTime, ErrorMessage: &message}); err != nil {
		t.Fatal(err)
	}
	checks, err := GetChecksByProject(db, project.ID, firstTime.Add(30*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) != 1 || checks[0].CheckedAt.Before(secondTime) || checks[0].ErrorMessage == nil {
		t.Fatalf("unexpected filtered checks: %+v", checks)
	}
}
