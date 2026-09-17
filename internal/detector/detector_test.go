package detector

import (
	"path/filepath"
	"testing"

	"github.com/hiroshiyoka/orpheus/internal/storage"
)

func TestConsecutiveFailures(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	p, err := storage.CreateProject(db, storage.Project{Name: "p", URL: "https://example.com", CheckIntervalSeconds: 60, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := storage.InsertCheck(db, storage.Check{ProjectID: p.ID, IsUp: false}); err != nil {
			t.Fatal(err)
		}
	}
	ok, err := ConsecutiveFailures(db, p.ID, 3)
	if err != nil || !ok {
		t.Fatalf("expected true, got %v err %v", ok, err)
	}
	if _, err := storage.InsertCheck(db, storage.Check{ProjectID: p.ID, IsUp: true}); err != nil {
		t.Fatal(err)
	}
	ok, err = ConsecutiveFailures(db, p.ID, 3)
	if err != nil || ok {
		t.Fatalf("expected false after success, got %v", ok)
	}
	ok, err = ConsecutiveFailures(db, p.ID, 5)
	if err != nil || ok {
		t.Fatalf("expected false when not enough rows")
	}
}

func TestCheckDowntime(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	p, err := storage.CreateProject(db, storage.Project{Name: "p", URL: "https://example.com", CheckIntervalSeconds: 60, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := storage.InsertCheck(db, storage.Check{ProjectID: p.ID, IsUp: false}); err != nil {
			t.Fatal(err)
		}
	}
	incident, err := CheckDowntime(db, p.ID, 3)
	if err != nil || incident == nil || incident.Type != "downtime" {
		t.Fatalf("expected incident, got %v err %v", incident, err)
	}
	dup, err := CheckDowntime(db, p.ID, 3)
	if err != nil || dup != nil {
		t.Fatalf("expected no duplicate, got %v", dup)
	}
	if _, err := storage.InsertCheck(db, storage.Check{ProjectID: p.ID, IsUp: true}); err != nil {
		t.Fatal(err)
	}
	none, err := CheckDowntime(db, p.ID, 3)
	if err != nil || none != nil {
		t.Fatalf("expected nil after success")
	}
}

func TestResolveDowntime(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	p, err := storage.CreateProject(db, storage.Project{Name: "p", URL: "https://example.com", CheckIntervalSeconds: 60, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := storage.InsertCheck(db, storage.Check{ProjectID: p.ID, IsUp: false}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := CheckDowntime(db, p.ID, 3); err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveDowntime(db, p.ID, true)
	if err != nil || resolved == nil || resolved.ResolvedAt == nil {
		t.Fatalf("expected resolved, got %v err %v", resolved, err)
	}
	active, err := storage.GetActiveIncidents(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 {
		t.Fatalf("expected no active incidents")
	}
	none, err := ResolveDowntime(db, p.ID, true)
	if err != nil || none != nil {
		t.Fatalf("expected nil when no active")
	}
	if _, err := storage.InsertCheck(db, storage.Check{ProjectID: p.ID, IsUp: false}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := storage.InsertCheck(db, storage.Check{ProjectID: p.ID, IsUp: false}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := CheckDowntime(db, p.ID, 3); err != nil {
		t.Fatal(err)
	}
	still, err := ResolveDowntime(db, p.ID, false)
	if err != nil || still != nil {
		t.Fatalf("expected nil when not up")
	}
}
