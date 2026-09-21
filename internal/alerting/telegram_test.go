package alerting

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendTelegramAlert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botTOKEN/sendMessage" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{`"chat_id":"CHAT"`, `"text":"down"`, `"parse_mode":"Markdown"`} {
			if !strings.Contains(string(body), want) {
				t.Errorf("body %q does not contain %q", body, want)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	originalTransport := http.DefaultTransport
	http.DefaultTransport = rewriteTransport{base: server.URL, transport: originalTransport}
	defer func() { http.DefaultTransport = originalTransport }()

	if err := SendTelegramAlert("TOKEN", "CHAT", "down"); err != nil {
		t.Fatal(err)
	}
}

type rewriteTransport struct {
	base      string
	transport http.RoundTripper
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = "http"
	clone.URL.Host = strings.TrimPrefix(t.base, "http://")
	return t.transport.RoundTrip(clone)
}
