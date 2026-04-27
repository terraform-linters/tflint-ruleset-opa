package opa

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/v1/bundle"
)

func testBundleBytes(t *testing.T) []byte {
	t.Helper()

	b := bundle.Bundle{
		Modules: []bundle.ModuleFile{
			{
				URL:    "/policy.rego",
				Path:   "/policy.rego",
				Raw:    []byte(`package tflint`),
				Parsed: nil,
			},
		},
		Data: map[string]interface{}{"foo": "bar"},
	}

	var buf bytes.Buffer
	if err := bundle.NewWriter(&buf).Write(b); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestFetchBundle(t *testing.T) {
	bundleBytes := testBundleBytes(t)

	t.Run("successful fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(bundleBytes)
		}))
		defer server.Close()

		b, err := fetchBundle(context.Background(), server.URL, "")
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Modules) == 0 {
			t.Fatal("expected at least one module")
		}
	})

	t.Run("bearer token", func(t *testing.T) {
		var gotAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			w.Write(bundleBytes)
		}))
		defer server.Close()

		t.Setenv("TFLINT_OPA_BUNDLE_TOKEN", "test-token-123")

		_, err := fetchBundle(context.Background(), server.URL, "")
		if err != nil {
			t.Fatal(err)
		}
		if gotAuth != "Bearer test-token-123" {
			t.Fatalf("expected Authorization header 'Bearer test-token-123', got '%s'", gotAuth)
		}
	})

	t.Run("no token", func(t *testing.T) {
		var gotAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			w.Write(bundleBytes)
		}))
		defer server.Close()

		_, err := fetchBundle(context.Background(), server.URL, "")
		if err != nil {
			t.Fatal(err)
		}
		if gotAuth != "" {
			t.Fatalf("expected no Authorization header, got '%s'", gotAuth)
		}
	})

	t.Run("HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		_, err := fetchBundle(context.Background(), server.URL, "")
		if err == nil {
			t.Fatal("expected error for HTTP 403")
		}
	})

	t.Run("bundle too large", func(t *testing.T) {
		original := maxBundleSize
		maxBundleSize = 10
		defer func() { maxBundleSize = original }()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(bundleBytes)
		}))
		defer server.Close()

		_, err := fetchBundle(context.Background(), server.URL, "")
		if err == nil {
			t.Fatal("expected error for oversized bundle")
		}
		if !strings.Contains(err.Error(), "exceeds maximum size (") {
			t.Fatalf("expected 'exceeds maximum size' error, got: %s", err)
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err := fetchBundle(context.Background(), "http://127.0.0.1:0/nonexistent", "")
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})

	t.Run("ETag caching", func(t *testing.T) {
		fetchCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("If-None-Match") == `"test-etag"` {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			fetchCount++
			w.Header().Set("ETag", `"test-etag"`)
			w.Write(bundleBytes)
		}))
		defer server.Close()

		cacheDir := t.TempDir()

		// First fetch — no cache, full download
		b, err := fetchBundle(context.Background(), server.URL, cacheDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Modules) == 0 {
			t.Fatal("expected at least one module")
		}
		if fetchCount != 1 {
			t.Fatalf("expected 1 fetch, got %d", fetchCount)
		}

		// Second fetch — should use cache (304)
		b, err = fetchBundle(context.Background(), server.URL, cacheDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Modules) == 0 {
			t.Fatal("expected at least one module from cache")
		}
		if fetchCount != 1 {
			t.Fatalf("expected no additional fetch, got %d total", fetchCount)
		}
	})

	t.Run("ETag cache miss on changed content", func(t *testing.T) {
		etag := `"v1"`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("If-None-Match") == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("ETag", etag)
			w.Write(bundleBytes)
		}))
		defer server.Close()

		cacheDir := t.TempDir()

		// First fetch
		_, err := fetchBundle(context.Background(), server.URL, cacheDir)
		if err != nil {
			t.Fatal(err)
		}

		// Server changes ETag — simulate content update
		etag = `"v2"`

		// Second fetch — old ETag doesn't match, full download
		b, err := fetchBundle(context.Background(), server.URL, cacheDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Modules) == 0 {
			t.Fatal("expected at least one module")
		}
	})

	t.Run("cache fallback on corrupted cache", func(t *testing.T) {
		fetchCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("If-None-Match") == `"test-etag"` {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			fetchCount++
			w.Header().Set("ETag", `"test-etag"`)
			w.Write(bundleBytes)
		}))
		defer server.Close()

		cacheDir := t.TempDir()

		// First fetch — populates cache
		_, err := fetchBundle(context.Background(), server.URL, cacheDir)
		if err != nil {
			t.Fatal(err)
		}
		if fetchCount != 1 {
			t.Fatalf("expected 1 fetch, got %d", fetchCount)
		}

		// Corrupt the cached bundle
		cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(server.URL)))
		os.WriteFile(filepath.Join(cacheDir, cacheKey+".tar.gz"), []byte("corrupted"), 0600)

		// Second fetch — 304, cache read fails, retries without cache
		b, err := fetchBundle(context.Background(), server.URL, cacheDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Modules) == 0 {
			t.Fatal("expected at least one module after cache fallback")
		}
		if fetchCount != 2 {
			t.Fatalf("expected 2 fetches after cache fallback, got %d", fetchCount)
		}
	})

	t.Run("no caching when cache dir is empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("ETag", `"some-etag"`)
			w.Write(bundleBytes)
		}))
		defer server.Close()

		b, err := fetchBundle(context.Background(), server.URL, "")
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Modules) == 0 {
			t.Fatal("expected at least one module")
		}
	})
}
