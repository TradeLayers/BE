package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"

	"github.com/TradeLayers/BE/internal/model"
)

// mockTokenVerifier is a simple mock for the TokenVerifier interface.
type mockTokenVerifier struct {
	verifyFn func(ctx context.Context, idToken string) (*auth.Token, error)
}

func (m *mockTokenVerifier) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return m.verifyFn(ctx, idToken)
}

func TestFirebaseAuth(t *testing.T) {
	tests := []struct {
		name             string
		authHeader       string
		mockVerifier     *mockTokenVerifier
		expectedStatus   int
		expectedContext  *model.UserContext
	}{
		{
			name:       "missing authorization header - returns 401",
			authHeader: "",
			mockVerifier: &mockTokenVerifier{
				verifyFn: func(ctx context.Context, idToken string) (*auth.Token, error) {
					return nil, nil
				},
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedContext: nil,
		},
		{
			name:       "invalid token - returns 401",
			authHeader: "Bearer invalid-token",
			mockVerifier: &mockTokenVerifier{
				verifyFn: func(ctx context.Context, idToken string) (*auth.Token, error) {
					return nil, errors.New("token is invalid")
				},
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedContext: nil,
		},
		{
			name:       "valid token with all claims",
			authHeader: "Bearer valid-token",
			mockVerifier: &mockTokenVerifier{
				verifyFn: func(ctx context.Context, idToken string) (*auth.Token, error) {
					return &auth.Token{
						UID: "user-123",
						Claims: map[string]interface{}{
							"email": "test@example.com",
							"name":  "Test User",
						},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedContext: &model.UserContext{
				FirebaseId: "user-123",
				Email:      "test@example.com",
				Name:       "Test User",
			},
		},
		{
			name:       "valid token with no claims - empty email and name",
			authHeader: "Bearer valid-token-no-claims",
			mockVerifier: &mockTokenVerifier{
				verifyFn: func(ctx context.Context, idToken string) (*auth.Token, error) {
					return &auth.Token{
						UID:    "user-456",
						Claims: map[string]interface{}{},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedContext: &model.UserContext{
				FirebaseId: "user-456",
				Email:      "",
				Name:       "",
			},
		},
		{
			name:       "valid token with email but no name",
			authHeader: "Bearer valid-token-email-only",
			mockVerifier: &mockTokenVerifier{
				verifyFn: func(ctx context.Context, idToken string) (*auth.Token, error) {
					return &auth.Token{
						UID: "user-789",
						Claims: map[string]interface{}{
							"email": "only-email@example.com",
						},
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedContext: &model.UserContext{
				FirebaseId: "user-789",
				Email:      "only-email@example.com",
				Name:       "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var capturedContext *model.UserContext
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			r := gin.New()
			r.Use(FirebaseAuth(tt.mockVerifier))
			r.GET("/protected", func(c *gin.Context) {
				if val, exists := c.Get("userContext"); exists {
					uc := val.(model.UserContext)
					capturedContext = &uc
				}
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Execute
			r.ServeHTTP(w, req)

			// Verify status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Verify user context
			if tt.expectedContext == nil && capturedContext != nil {
				t.Error("expected no user context but got one")
			}
			if tt.expectedContext != nil {
				if capturedContext == nil {
					t.Fatal("expected user context but got nil")
				}
				if capturedContext.FirebaseId != tt.expectedContext.FirebaseId {
					t.Errorf("expected FirebaseId %s, got %s", tt.expectedContext.FirebaseId, capturedContext.FirebaseId)
				}
				if capturedContext.Email != tt.expectedContext.Email {
					t.Errorf("expected Email %s, got %s", tt.expectedContext.Email, capturedContext.Email)
				}
				if capturedContext.Name != tt.expectedContext.Name {
					t.Errorf("expected Name %s, got %s", tt.expectedContext.Name, capturedContext.Name)
				}
			}
		})
	}
}
