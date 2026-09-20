package http_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	controller "summary-your-usage/internal/controller/http"
	"summary-your-usage/internal/entity"
	"summary-your-usage/internal/repo/persistent"
	"summary-your-usage/internal/usecase"
	"summary-your-usage/migrations"
	"summary-your-usage/pkg/database"
)

func openAPI(t *testing.T, path string) (http.Handler, *sql.DB) {
	t.Helper()
	db, err := database.Open(t.Context(), "sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	provider, err := migrations.New(db, "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(t.Context()); err != nil {
		t.Fatal(err)
	}
	return controller.NewRouter(usecase.NewUser(persistent.NewUser(db))), db
}

func request(t *testing.T, handler http.Handler, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("%s %s: status = %d, want %d; body = %s", method, path, response.Code, status, response.Body)
	}
	return response
}

func decode[T any](t *testing.T, response *httptest.ResponseRecorder) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
		t.Fatalf("decode %s: %v", response.Body, err)
	}
	return value
}

func assertError(t *testing.T, response *httptest.ResponseRecorder, code string) {
	t.Helper()
	value := decode[struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}](t, response)
	if value.Error.Code != code || value.Error.Message == "" {
		t.Fatalf("unexpected error response: %s; want code %q and a message", response.Body, code)
	}
}

func TestUserCRUDAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "users.db")
	api, db := openAPI(t, path)
	empty := decode[[]entity.User](t, request(t, api, "GET", "/api/v1/users", "", 200))
	if empty == nil || len(empty) != 0 {
		t.Fatalf("unexpected empty list: %+v", empty)
	}
	created := request(t, api, "POST", "/api/v1/users", `{"name":"  张三  ","email":"  Alice@Example.COM  "}`, 201)
	user := decode[entity.User](t, created)
	location := fmt.Sprintf("/api/v1/users/%d", user.ID)
	if user.ID < 1 || user.Name != "张三" || user.Email != "alice@example.com" || user.CreatedAt.IsZero() || !user.CreatedAt.Equal(user.UpdatedAt) || created.Header().Get("Location") != location {
		t.Fatalf("unexpected created user: %+v, Location: %s", user, created.Header().Get("Location"))
	}
	assertError(t, request(t, api, "POST", "/api/v1/users", `{"name":"重复","email":"ALICE@example.com"}`, 409), "email_exists")
	other := decode[entity.User](t, request(t, api, "POST", "/api/v1/users", `{"name":"李四","email":"bob@example.com"}`, 201))
	assertError(t, request(t, api, "PUT", location, `{"name":"不应保存","email":"BOB@example.com"}`, 409), "email_exists")
	unchanged := decode[entity.User](t, request(t, api, "GET", location, "", 200))
	if unchanged != user {
		t.Fatalf("failed update changed user: %+v, want %+v", unchanged, user)
	}
	updated := decode[entity.User](t, request(t, api, "PUT", location, `{"name":"  王五  ","email":"  NEW@example.com  "}`, 200))
	if updated.ID != user.ID || updated.Name != "王五" || updated.Email != "new@example.com" || !updated.CreatedAt.Equal(user.CreatedAt) || updated.UpdatedAt.Before(user.UpdatedAt) {
		t.Fatalf("unexpected updated user: %+v", updated)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	api, _ = openAPI(t, path)
	persisted := decode[entity.User](t, request(t, api, "GET", location, "", 200))
	if persisted != updated {
		t.Fatalf("reopened database: %+v, want %+v", persisted, updated)
	}
	deleted := request(t, api, "DELETE", location, "", 204)
	if deleted.Body.Len() != 0 {
		t.Fatalf("delete response must have an empty body: %s", deleted.Body)
	}
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		assertError(t, request(t, api, method, location, `{"name":"不存在","email":"missing@example.com"}`, 404), "user_not_found")
	}
	remaining := decode[[]entity.User](t, request(t, api, "GET", "/api/v1/users", "", 200))
	if len(remaining) != 1 || remaining[0].ID != other.ID {
		t.Fatalf("unexpected remaining users: %+v", remaining)
	}
}

func TestInvalidRequests(t *testing.T) {
	api, _ := openAPI(t, filepath.Join(t.TempDir(), "users.db"))
	for _, tc := range []struct {
		name, method, path, body, code string
		status                         int
	}{
		{"missing fields", "POST", "/api/v1/users", `{}`, "validation_error", 400},
		{"blank name", "POST", "/api/v1/users", `{"name":"  ","email":"a@example.com"}`, "validation_error", 400},
		{"long name", "POST", "/api/v1/users", `{"name":"` + strings.Repeat("张", 101) + `","email":"a@example.com"}`, "validation_error", 400},
		{"invalid email", "POST", "/api/v1/users", `{"name":"Alice","email":"invalid"}`, "validation_error", 400},
		{"empty JSON", "POST", "/api/v1/users", "", "invalid_json", 400},
		{"malformed JSON", "POST", "/api/v1/users", `{`, "invalid_json", 400},
		{"unknown field", "POST", "/api/v1/users", `{"name":"Alice","email":"a@example.com","admin":true}`, "invalid_json", 400},
		{"wrong type", "POST", "/api/v1/users", `{"name":42,"email":"a@example.com"}`, "invalid_json", 400},
		{"multiple values", "POST", "/api/v1/users", `{"name":"Alice","email":"a@example.com"} {}`, "invalid_json", 400},
		{"trailing garbage", "POST", "/api/v1/users", `{"name":"Alice","email":"a@example.com"} invalid`, "invalid_json", 400},
		{"oversized body", "POST", "/api/v1/users", `{"name":"` + strings.Repeat("a", 1<<20) + `","email":"a@example.com"}`, "body_too_large", 413},
		{"invalid update", "PUT", "/api/v1/users/1", `{}`, "validation_error", 400},
		{"zero ID", "GET", "/api/v1/users/0", "", "validation_error", 400},
		{"negative ID", "DELETE", "/api/v1/users/-1", "", "validation_error", 400},
		{"invalid ID", "GET", "/api/v1/users/abc", "", "validation_error", 400},
		{"overflow ID", "GET", "/api/v1/users/9223372036854775808", "", "validation_error", 400},
		{"unknown route", "GET", "/api/v1/missing", "", "not_found", 404},
		{"unsupported method", "PATCH", "/api/v1/users/1", "", "method_not_allowed", 405},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertError(t, request(t, api, tc.method, tc.path, tc.body, tc.status), tc.code)
		})
	}
	t.Run("content type", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(`{"name":"Alice","email":"a@example.com"}`))
		req.Header.Set("Content-Type", "text/plain")
		response := httptest.NewRecorder()
		api.ServeHTTP(response, req)
		if response.Code != 415 {
			t.Fatalf("status = %d, want 415", response.Code)
		}
		assertError(t, response, "unsupported_media_type")
	})
	users := decode[[]entity.User](t, request(t, api, "GET", "/api/v1/users", "", 200))
	if len(users) != 0 {
		t.Fatalf("invalid requests created users: %+v", users)
	}
}
