package handler_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/aplulu/hakoniwa/internal/interface/http/handler"
)

func TestGatewayHandler_RedirectToDashboard_CacheControl(t *testing.T) {
	// 1. Setup temporary static directory with index.html
	tempDir, err := os.MkdirTemp("", "static")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexContent := "<html><body>Index</body></html>"
	if err := os.WriteFile(filepath.Join(tempDir, "index.html"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write index.html: %v", err)
	}

	// 2. Create GatewayHandler with minimal dependencies
	// We use nil for dependencies that are not expected to be called in this scenario
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewGatewayHandler(
		nil, // authUsecase
		nil, // instanceUsecase
		nil, // apiServer
		nil, // proxyHandler
		tempDir,
		logger,
	)

	// 3. Create a request that triggers redirectToDashboard (Unauthenticated, not API, not static asset)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// 4. Execute
	h.ServeHTTP(w, req)

	// 5. Verify Headers
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != "/_hakoniwa/" {
		t.Errorf("Expected Location /_hakoniwa/, got %q", location)
	}

	cacheControl := resp.Header.Get("Cache-Control")
	expectedCacheControl := "no-store, no-cache, must-revalidate, proxy-revalidate"
	if cacheControl != expectedCacheControl {
		t.Errorf("Expected Cache-Control %q, got %q", expectedCacheControl, cacheControl)
	}

	pragma := resp.Header.Get("Pragma")
	if pragma != "no-cache" {
		t.Errorf("Expected Pragma no-cache, got %q", pragma)
	}

	expires := resp.Header.Get("Expires")
	if expires != "0" {
		t.Errorf("Expected Expires 0, got %q", expires)
	}
}
