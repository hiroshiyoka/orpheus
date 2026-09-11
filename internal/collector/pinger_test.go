package collector

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ok.Close()
	fail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer fail.Close()
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer slow.Close()

	good := Ping(ok.URL, 2*time.Second)
	if !good.IsUp || good.StatusCode == nil || *good.StatusCode != http.StatusOK || good.ErrorMessage != nil {
		t.Fatalf("unexpected success result: %+v", good)
	}
	bad := Ping(fail.URL, 2*time.Second)
	if bad.IsUp || bad.StatusCode == nil || *bad.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected failure result: %+v", bad)
	}
	expired := Ping(slow.URL, 50*time.Millisecond)
	if expired.IsUp || expired.ErrorMessage == nil {
		t.Fatalf("unexpected timeout result: %+v", expired)
	}
}
