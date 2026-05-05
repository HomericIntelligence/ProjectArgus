package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// base is a shared helper for HTTP-based pollers.
type base struct {
	name   string
	client *http.Client
}

// getJSON performs a GET request to url and JSON-decodes the response body into dst.
// It returns an error if the request fails or the status code is not 200.
func (b *base) getJSON(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
