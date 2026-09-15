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
