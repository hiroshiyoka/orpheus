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
		Start(ctx, db, []storage.Project{p}, fakePinger, 3, "", "", nil)
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

func TestStartSendsIncidentAlerts(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "orpheus.db"), filepath.Join("..", "..", "migrations", "0001_init.sql"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p, err := storage.CreateProject(db, storage.Project{Name: "Example", URL: "https://example.com", CheckIntervalSeconds: 1, IsActive: true})
	if err != nil {
		t.Fatal(err)
	}

	var checks int32
	fakePinger := func(url string, timeout time.Duration) storage.Check {
		up := atomic.AddInt32(&checks, 1) > 3
		return storage.Check{IsUp: up}
	}
	alerts := make(chan string, 2)
	fakeSender := func(botToken, chatID, message string) error {
		alerts <- message
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		Start(ctx, db, []storage.Project{p}, fakePinger, 3, "token", "chat", fakeSender)
		close(done)
	}()

	select {
	case message := <-alerts:
		if message != "[Example] Down\nURL: https://example.com" {
			t.Fatalf("unexpected downtime alert: %q", message)
		}
	case <-time.After(5 * time.Second):
		cancel()
		<-done
		t.Fatal("timed out waiting for downtime alert")
	}
	select {
	case message := <-alerts:
		if message != "[Example] Recovered\nURL: https://example.com" {
			t.Fatalf("unexpected recovery alert: %q", message)
		}
	case <-time.After(2 * time.Second):
		cancel()
		<-done
		t.Fatal("timed out waiting for recovery alert")
	}
	var logged int
	if err := db.QueryRow(`SELECT COUNT(*) FROM alerts_sent WHERE incident_id = 1`).Scan(&logged); err != nil {
		t.Fatal(err)
	}
	if logged != 2 {
		t.Fatalf("expected two logged alerts, got %d", logged)
	}
	cancel()
	<-done
}

func ptrInt(v int) *int { return &v }
