package collector

import (
	"net/http"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/storage"
)

func Ping(url string, timeout time.Duration) storage.Check {
	start := time.Now()
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	elapsed := int(time.Since(start).Milliseconds())
	if err != nil {
		message := err.Error()
		return storage.Check{ResponseTimeMS: &elapsed, ErrorMessage: &message}
	}
	defer resp.Body.Close()
	code := resp.StatusCode
	return storage.Check{IsUp: code < 400, StatusCode: &code, ResponseTimeMS: &elapsed}
}
