package dataurl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A function instance handles many requests and shares a small pool of file
// descriptors, so a response body that is never closed eventually breaks the
// whole process.
func TestConvertClosesTheResponseBody(t *testing.T) {
	closed := false

	original := DefaultHTTPClient
	DefaultHTTPClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp, err := original.Do(req)
		if err != nil {
			return nil, err
		}
		resp.Body = &closeTracker{ReadCloser: resp.Body, onClose: func() { closed = true }}
		return resp, nil
	})}
	defer func() { DefaultHTTPClient = original }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png"))
	}))
	defer srv.Close()

	if _, err := Convert(context.Background(), srv.URL, nil); err != nil {
		t.Fatalf("Convert returned an error: %v", err)
	}
	if !closed {
		t.Fatal("Convert must close the response body")
	}
}

func TestConvertFailsOnNonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := Convert(context.Background(), srv.URL, nil); err == nil {
		t.Fatal("expected an error for a 404 avatar response")
	}
}

func TestDefaultHTTPClientHasATimeout(t *testing.T) {
	if DefaultHTTPClient.Timeout == 0 {
		t.Fatal("DefaultHTTPClient needs a timeout: a hanging avatar host would otherwise pin the request")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type closeTracker struct {
	ReadCloser interface {
		Read([]byte) (int, error)
		Close() error
	}
	onClose func()
}

func (c *closeTracker) Read(p []byte) (int, error) { return c.ReadCloser.Read(p) }
func (c *closeTracker) Close() error {
	c.onClose()
	return c.ReadCloser.Close()
}
