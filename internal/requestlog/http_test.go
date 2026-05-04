package requestlog

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestHTTPMiddlewareAddsRequestIDToResponseAndLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, observed := observer.New(zap.InfoLevel)
	baseLog := zap.New(core)

	var ginRequestID string
	var contextRequestID string

	router := gin.New()
	router.Use(HTTPMiddleware(baseLog))
	router.GET("/test", func(c *gin.Context) {
		value, _ := c.Get(RequestIDKey)
		ginRequestID, _ = value.(string)
		contextRequestID = RequestIDFromContext(c.Request.Context())

		FromContext(c.Request.Context()).Info("handler_log")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	headerRequestID := recorder.Header().Get(RequestIDHeader)
	if headerRequestID == "" {
		t.Fatal("expected X-Request-Id response header to be set")
	}
	if _, err := uuid.Parse(headerRequestID); err != nil {
		t.Fatalf("expected request ID to be a UUID, got %q: %v", headerRequestID, err)
	}
	if ginRequestID != headerRequestID {
		t.Fatalf("expected gin context request ID %q to match response header %q", ginRequestID, headerRequestID)
	}
	if contextRequestID != headerRequestID {
		t.Fatalf("expected request context request ID %q to match response header %q", contextRequestID, headerRequestID)
	}

	entries := observed.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(entries))
	}

	for _, entry := range entries {
		requestID, _ := entry.ContextMap()[RequestIDKey].(string)
		if requestID != headerRequestID {
			t.Fatalf("expected log request ID %q, got %q for message %q", headerRequestID, requestID, entry.Message)
		}
	}
}

func TestHTTPMiddlewareGeneratesUniqueRequestIDsForParallelRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, observed := observer.New(zap.InfoLevel)
	baseLog := zap.New(core)

	router := gin.New()
	router.Use(HTTPMiddleware(baseLog))
	router.GET("/test", func(c *gin.Context) {
		FromContext(c.Request.Context()).Info("handler_log")
		c.Status(http.StatusNoContent)
	})

	const requestCount = 12

	headers := make([]string, requestCount)
	var wg sync.WaitGroup
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			headers[idx] = recorder.Header().Get(RequestIDHeader)
		}(i)
	}
	wg.Wait()

	seenHeaders := make(map[string]struct{}, requestCount)
	for _, header := range headers {
		if header == "" {
			t.Fatal("expected every response to include X-Request-Id")
		}
		if _, err := uuid.Parse(header); err != nil {
			t.Fatalf("expected request ID to be a UUID, got %q: %v", header, err)
		}
		if _, exists := seenHeaders[header]; exists {
			t.Fatalf("duplicate request ID generated: %q", header)
		}
		seenHeaders[header] = struct{}{}
	}

	handlerEntries := observed.FilterMessage("handler_log").All()
	if len(handlerEntries) != requestCount {
		t.Fatalf("expected %d handler logs, got %d", requestCount, len(handlerEntries))
	}

	seenLogIDs := make(map[string]struct{}, requestCount)
	for _, entry := range handlerEntries {
		requestID, _ := entry.ContextMap()[RequestIDKey].(string)
		if requestID == "" {
			t.Fatalf("expected handler log %q to include request_id", entry.Message)
		}
		seenLogIDs[requestID] = struct{}{}
	}

	if len(seenLogIDs) != requestCount {
		t.Fatalf("expected %d unique request IDs in handler logs, got %d", requestCount, len(seenLogIDs))
	}

	for header := range seenHeaders {
		if _, exists := seenLogIDs[header]; !exists {
			t.Fatalf("response header request ID %q was not found in handler logs", header)
		}
	}
}
