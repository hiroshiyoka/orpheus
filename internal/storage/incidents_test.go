package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestIncidentStorage(t *testing.T) {
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
	description := "error spike detected"
	first, err := CreateIncident(db, Incident{
		ProjectID:   project.ID,
		Type:        "error_spike",
		StartedAt:   start,
		Description: &description,
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == 0 || first.ResolvedAt != nil || first.Description == nil || *first.Description != description {
		t.Fatalf("unexpected created incident: %+v", first)
	}

	second, err := CreateIncident(db, Incident{
		ProjectID: project.ID,
		Type:      "downtime",
		StartedAt: start.Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	active, err := GetActiveIncidents(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 2 {
		t.Fatalf("unexpected active incidents: %+v", active)
	}

	resolvedAt := start.Add(2 * time.Hour)
	if err := ResolveIncident(db, first.ID, resolvedAt); err != nil {
		t.Fatal(err)
	}
	active, err = GetActiveIncidents(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].ID != second.ID {
		t.Fatalf("unexpected active incidents after resolve: %+v", active)
	}

	incidents, err := ListIncidents(db, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(incidents) != 2 || incidents[0].ID != first.ID || incidents[0].ResolvedAt == nil {
		t.Fatalf("unexpected incident list: %+v", incidents)
	}
}
