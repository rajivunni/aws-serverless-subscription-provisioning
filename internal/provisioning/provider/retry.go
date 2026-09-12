package provider

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type retryRoundTripper struct {
	rt         http.RoundTripper
	maxRetries int
	baseDelay  time.Duration
}

func (r *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			delay := r.baseDelay * time.Duration(attempt)
			log.Printf("Provider retry status update")
			time.Sleep(delay)
		}

		resp, err := r.rt.RoundTrip(req)
		if err != nil {
			lastErr = err
			log.Printf("Provider retry operation failed; details omitted")
			continue
		}

		// Retry on 5xx server errors
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			log.Printf("Provider retry status update")
			continue
		}

		return resp, nil
	}
	return nil, fmt.Errorf("all %d attempts failed: %w", r.maxRetries+1, lastErr)
}
