package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/hiroshiyoka/orpheus/internal/storage"
)

func TestHealth(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := NewRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
}
