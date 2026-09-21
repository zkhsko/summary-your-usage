package ui_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"summary-your-usage/ui"
)

func TestHandler(t *testing.T) {
	router := chi.NewRouter()
	router.Mount("/ui", http.StripPrefix("/ui", ui.Handler()))
	request := func(method, target string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(method, target, nil))
		return response
	}

	index := request(http.MethodGet, "/ui/")
	if index.Code != http.StatusOK || !strings.Contains(index.Body.String(), `<div id="app">`) {
		t.Fatalf("unexpected index response: %d %s", index.Code, index.Body)
	}
	for _, target := range []string{
		"/ui/users", "/ui/groups", "/ui/dashboard/overview", "/ui/users/", "/ui/users?search=alice", "/ui/missing",
	} {
		t.Run(target, func(t *testing.T) {
			response := request(http.MethodGet, target)
			if response.Code != http.StatusOK || response.Body.String() != index.Body.String() ||
				!strings.HasPrefix(response.Header().Get("Content-Type"), "text/html") {
				t.Fatalf("expected app HTML, got %d %s", response.Code, response.Body)
			}
		})
	}

	t.Run("HEAD page", func(t *testing.T) {
		response := request(http.MethodHead, "/ui/users")
		if response.Code != http.StatusOK || response.Body.Len() != 0 ||
			response.Header().Get("Content-Length") != index.Header().Get("Content-Length") {
			t.Fatalf("unexpected HEAD response: %d %v %s", response.Code, response.Header(), response.Body)
		}
	})

	t.Run("static assets", func(t *testing.T) {
		assets := regexp.MustCompile(`(?:src|href)="(/ui/assets/[^"]+)"`).FindAllStringSubmatch(index.Body.String(), -1)
		if len(assets) == 0 {
			t.Fatal("index has no static assets")
		}
		for _, asset := range assets {
			response := request(http.MethodGet, asset[1])
			if response.Code != http.StatusOK || response.Body.Len() == 0 ||
				strings.HasPrefix(response.Header().Get("Content-Type"), "text/html") {
				t.Fatalf("unexpected asset response for %s: %d %v", asset[1], response.Code, response.Header())
			}
		}
	})

	for _, target := range []string{
		"/ui/assets/missing.js", "/ui/assets/missing.css", "/ui/assets/missing", "/ui/missing.svg", "/ui-api/missing",
	} {
		t.Run(target, func(t *testing.T) {
			response := request(http.MethodGet, target)
			if response.Code != http.StatusNotFound {
				t.Fatalf("expected 404, got %d %s", response.Code, response.Body)
			}
		})
	}

	t.Run("POST page", func(t *testing.T) {
		response := request(http.MethodPost, "/ui/users")
		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d %s", response.Code, response.Body)
		}
	})
}
