package router

import "testing"

func TestBuildAllowedOrigins(t *testing.T) {
	t.Run("supports multiple configured origins", func(t *testing.T) {
		origins := buildAllowedOrigins("https://app.example.com,https://admin.example.com")

		assertContainsOrigin(t, origins, "https://app.example.com")
		assertContainsOrigin(t, origins, "https://admin.example.com")
		assertDoesNotContainOrigin(t, origins, "http://localhost:3001")
	})

	t.Run("adds local frontend ports when localhost is configured", func(t *testing.T) {
		origins := buildAllowedOrigins("http://localhost:3000")

		assertContainsOrigin(t, origins, "http://localhost:3000")
		assertContainsOrigin(t, origins, "http://localhost:3001")
		assertContainsOrigin(t, origins, "http://localhost:3002")
		assertContainsOrigin(t, origins, "http://127.0.0.1:3001")
	})

	t.Run("defaults to local frontend ports when env is empty", func(t *testing.T) {
		origins := buildAllowedOrigins("")

		assertContainsOrigin(t, origins, "http://localhost:3000")
		assertContainsOrigin(t, origins, "http://localhost:3001")
		assertContainsOrigin(t, origins, "http://localhost:3002")
	})
}

func assertContainsOrigin(t *testing.T, origins []string, expectedOrigin string) {
	t.Helper()

	if !isAllowedOrigin(expectedOrigin, origins) {
		t.Fatalf("expected origin %q to be allowed, got %v", expectedOrigin, origins)
	}
}

func assertDoesNotContainOrigin(t *testing.T, origins []string, unexpectedOrigin string) {
	t.Helper()

	if isAllowedOrigin(unexpectedOrigin, origins) {
		t.Fatalf("expected origin %q to be blocked, got %v", unexpectedOrigin, origins)
	}
}
