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

func TestStart(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	p, err := storage.CreateProject(db, storage.Project{Name: "Example", URL: srv.URL, CheckIntervalSeconds: 1, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}

	var count int32
	fakePinger := func(url string, timeout time.Duration) storage.Check {
		atomic.AddInt32(&count, 1)
		return storage.Check{IsUp: true, StatusCode: ptrInt(200)}
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		Start(ctx, db, []storage.Project{p}, fakePinger, 3)
		close(done)
	}()

	time.Sleep(3 * time.Second)
	cancel()
	<-done

	checks, err := storage.GetChecksByProject(db, p.ID, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) == 0 {
		t.Fatal("expected at least one check row")
	}
}

func ptrInt(v int) *int { return &v }
