package checker

import (
	"context"
	"net/http"
	"time"

	"links-checker/internal/domain"
)

type LinkChecker struct {
	client *http.Client
}

func New() *LinkChecker {
	return &LinkChecker{
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c *LinkChecker) Check(ctx context.Context, url string) domain.LinkStatus {
	if url == "" {
		return domain.StatusNotAvailable
	}

	if url[:4] != "http" {
		url = "http://" + url
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.StatusNotAvailable
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.StatusNotAvailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest {
		return domain.StatusAvailable
	}

	return domain.StatusNotAvailable
}
