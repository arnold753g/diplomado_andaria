package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSAllowsAgencyAttractionAndFavoriteMethods(t *testing.T) {
	handler := CORS([]string{"http://127.0.0.1:3000"}).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	for _, method := range []string{"PUT", "DELETE"} {
		for _, origin := range []string{"http://127.0.0.1:3000", "https://untrusted.example"} {
			t.Run(method+origin, func(t *testing.T) {
				r := httptest.NewRequest(http.MethodOptions, "/api/v1/managed-attractions/1", nil)
				r.Header.Set("Origin", origin)
				r.Header.Set("Access-Control-Request-Method", method)
				r.Header.Set("Access-Control-Request-Headers", "content-type,x-csrf-token")
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)
				if origin == "https://untrusted.example" {
					if w.Header().Get("Access-Control-Allow-Origin") != "" {
						t.Fatal("untrusted origin allowed")
					}
					return
				}
				if w.Header().Get("Access-Control-Allow-Origin") != origin || w.Header().Get("Access-Control-Allow-Methods") != method || w.Header().Get("Access-Control-Allow-Credentials") != "true" {
					t.Fatalf("preflight denied: %v", w.Header())
				}
				if !strings.Contains(strings.ToLower(w.Header().Get("Access-Control-Allow-Headers")), "x-csrf-token") {
					t.Fatal("CSRF header not allowed")
				}
			})
		}
	}
}
