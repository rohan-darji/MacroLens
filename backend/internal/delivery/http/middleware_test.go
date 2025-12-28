package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsAllowedOrigin(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		want           bool
	}{
		{
			name:           "exact match",
			origin:         "chrome-extension://abcdefg12345",
			allowedOrigins: []string{"chrome-extension://abcdefg12345"},
			want:           true,
		},
		{
			name:           "wildcard match",
			origin:         "chrome-extension://abcdefg12345",
			allowedOrigins: []string{"chrome-extension://*"},
			want:           true,
		},
		{
			name:           "multiple allowed origins - matches first",
			origin:         "chrome-extension://abcdefg12345",
			allowedOrigins: []string{"chrome-extension://*", "http://localhost:3000"},
			want:           true,
		},
		{
			name:           "multiple allowed origins - matches second",
			origin:         "http://localhost:3000",
			allowedOrigins: []string{"chrome-extension://*", "http://localhost:3000"},
			want:           true,
		},
		{
			name:           "no match",
			origin:         "http://evil.com",
			allowedOrigins: []string{"chrome-extension://*"},
			want:           false,
		},
		{
			name:           "empty origin",
			origin:         "",
			allowedOrigins: []string{"chrome-extension://*"},
			want:           false,
		},
		{
			name:           "empty allowed list",
			origin:         "chrome-extension://abcdefg12345",
			allowedOrigins: []string{},
			want:           false,
		},
		{
			name:           "partial wildcard match",
			origin:         "chrome-extension://abcdefg12345",
			allowedOrigins: []string{"chrome-*"},
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedOrigin(tt.origin, tt.allowedOrigins)
			if got != tt.want {
				t.Errorf("isAllowedOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		method         string
		wantStatus     int
		checkHeaders   bool
		wantCORS       bool
	}{
		{
			name:           "allowed origin - GET request",
			origin:         "chrome-extension://abcdefg12345",
			allowedOrigins: []string{"chrome-extension://*"},
			method:         "GET",
			wantStatus:     http.StatusOK,
			checkHeaders:   true,
			wantCORS:       true,
		},
		{
			name:           "allowed origin - OPTIONS request",
			origin:         "chrome-extension://abcdefg12345",
			allowedOrigins: []string{"chrome-extension://*"},
			method:         "OPTIONS",
			wantStatus:     http.StatusNoContent,
			checkHeaders:   true,
			wantCORS:       true,
		},
		{
			name:           "disallowed origin",
			origin:         "http://evil.com",
			allowedOrigins: []string{"chrome-extension://*"},
			method:         "GET",
			wantStatus:     http.StatusOK,
			checkHeaders:   true,
			wantCORS:       false,
		},
		{
			name:           "no origin header",
			origin:         "",
			allowedOrigins: []string{"chrome-extension://*"},
			method:         "GET",
			wantStatus:     http.StatusOK,
			checkHeaders:   true,
			wantCORS:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			router.Use(CORSMiddleware(tt.allowedOrigins))
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			// Create request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
			}

			// Check CORS headers
			if tt.checkHeaders {
				corsHeader := w.Header().Get("Access-Control-Allow-Origin")
				if tt.wantCORS {
					if corsHeader != tt.origin {
						t.Errorf("Access-Control-Allow-Origin = %s, want %s", corsHeader, tt.origin)
					}
					if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
						t.Errorf("Access-Control-Allow-Credentials not set to true")
					}
				} else {
					if corsHeader != "" {
						t.Errorf("Access-Control-Allow-Origin should not be set for disallowed origin, got %s", corsHeader)
					}
				}
			}
		})
	}
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORSMiddleware([]string{"chrome-extension://*"}))
	router.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Create preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "chrome-extension://abcdefg12345")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 204 No Content
	if w.Code != http.StatusNoContent {
		t.Errorf("Preflight status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "chrome-extension://abcdefg12345" {
		t.Errorf("Access-Control-Allow-Origin not set correctly")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("Access-Control-Allow-Methods not set")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Errorf("Access-Control-Allow-Headers not set")
	}
	if w.Header().Get("Access-Control-Max-Age") == "" {
		t.Errorf("Access-Control-Max-Age not set")
	}
}
