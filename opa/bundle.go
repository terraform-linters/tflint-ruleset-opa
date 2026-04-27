package opa

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/open-policy-agent/opa/v1/bundle"
)

var bundleHTTPClient = &http.Client{Timeout: 30 * time.Second}

var maxBundleSize int64 = 256 << 20 // 256 MB

type limitedReader struct {
	r         io.Reader
	remaining int64
}

func (l *limitedReader) Read(p []byte) (int, error) {
	if l.remaining <= 0 {
		return 0, fmt.Errorf("bundle exceeds maximum size (%d MB)", maxBundleSize>>20)
	}
	if int64(len(p)) > l.remaining {
		p = p[:l.remaining]
	}
	n, err := l.r.Read(p)
	l.remaining -= int64(n)
	return n, err
}

func fetchBundle(ctx context.Context, bundleURL string, cacheDir string) (*bundle.Bundle, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", bundleURL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid bundle URL: %w", err)
	}

	if token := os.Getenv("TFLINT_OPA_BUNDLE_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(bundleURL)))
	if cacheDir != "" {
		if etag, err := os.ReadFile(filepath.Join(cacheDir, cacheKey+".etag")); err == nil {
			req.Header.Set("If-None-Match", string(etag))
		}
	}

	resp, err := bundleHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bundle from %s: %w", bundleURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified && cacheDir != "" {
		b, err := readCachedBundle(filepath.Join(cacheDir, cacheKey+".tar.gz"))
		if err == nil {
			return b, nil
		}
		os.Remove(filepath.Join(cacheDir, cacheKey+".etag"))
		return fetchBundle(ctx, bundleURL, "")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch bundle from %s: HTTP %d", bundleURL, resp.StatusCode)
	}

	lr := &limitedReader{r: resp.Body, remaining: maxBundleSize}
	body, err := io.ReadAll(lr)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle from %s: %w", bundleURL, err)
	}

	b, err := bundle.NewReader(bytes.NewReader(body)).Read()
	if err != nil {
		return nil, fmt.Errorf("failed to parse bundle from %s: %w", bundleURL, err)
	}

	if cacheDir != "" {
		if etag := resp.Header.Get("ETag"); etag != "" {
			writeCache(cacheDir, cacheKey, body, etag)
		}
	}

	return &b, nil
}

func readCachedBundle(path string) (*bundle.Bundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read cached bundle: %w", err)
	}
	b, err := bundle.NewReader(bytes.NewReader(data)).Read()
	if err != nil {
		return nil, fmt.Errorf("failed to parse cached bundle: %w", err)
	}
	return &b, nil
}

func writeCache(cacheDir, cacheKey string, bundleData []byte, etag string) {
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return
	}
	// Write bundle first, then etag — etag presence signals cache validity
	if err := os.WriteFile(filepath.Join(cacheDir, cacheKey+".tar.gz"), bundleData, 0600); err != nil {
		return
	}
	os.WriteFile(filepath.Join(cacheDir, cacheKey+".etag"), []byte(etag), 0600)
}
