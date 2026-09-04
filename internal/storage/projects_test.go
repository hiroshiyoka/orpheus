package storage

import (
	"path/filepath"
	"testing"
)

func TestProjectCRUD(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	zoneID := "zone-1"
	created, err := CreateProject(db, Project{
		Name:                 "Orpheus",
		URL:                  "https://example.com",
		CloudflareZoneID:     &zoneID,
		CheckIntervalSeconds: 30,
		IsActive:             true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.CreatedAt.IsZero() {
		t.Fatalf("created project has incomplete metadata: %+v", created)
	}

	got, err := GetProject(db, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != created.Name || got.URL != created.URL || got.CloudflareZoneID == nil || *got.CloudflareZoneID != zoneID {
		t.Fatalf("got project does not match created project: %+v", got)
	}

	got.Name = "Updated Orpheus"
	got.IsActive = false
	if err := UpdateProject(db, got); err != nil {
		t.Fatal(err)
	}
	updated, err := GetProject(db, got.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != got.Name || updated.IsActive != got.IsActive {
		t.Fatalf("project was not updated: %+v", updated)
	}

	projects, err := ListProjects(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].ID != created.ID {
		t.Fatalf("unexpected project list: %+v", projects)
	}
}
