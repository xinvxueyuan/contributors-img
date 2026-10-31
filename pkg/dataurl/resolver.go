package dataurl

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// DefaultHTTPClient fetches avatars. The timeout matters in this deployment:
// one render fetches up to 100 avatars in parallel, and a slow avatar host
// would otherwise hold the request until the platform kills it.
var DefaultHTTPClient = &http.Client{
	Timeout:   10 * time.Second,
	Transport: otelhttp.NewTransport(http.DefaultTransport),
}

func Convert(c context.Context, remoteURL string, extraParams map[string]string) (string, error) {
	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse avatar URL %q: %w", remoteURL, err)
	}
	q := u.Query()
	for k, v := range extraParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(c, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := DefaultHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch avatar %s: status %d", u.Redacted(), resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(body)), nil
}
