package http_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"summary-your-usage/internal/entity"
)

func TestGroupCRUDAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "groups.db")
	api, db := openAPI(t, path)
	empty := decode[[]entity.Group](t, request(t, api, "GET", "/ui-api/groups", "", 200))
	if empty == nil || len(empty) != 0 {
		t.Fatalf("unexpected empty list: %+v", empty)
	}
	group := decode[entity.Group](t, request(t, api, "POST", "/ui-api/groups",
		`{"name":"  标准组  ","billing_multiplier":1.25,"visible_other_group":true,"description":"  分组说明\n第二行  "}`, 201))
	groupPath := fmt.Sprintf("/ui-api/groups/%d", group.Id)
	if group.Id < 1 || group.Name != "标准组" || group.BillingMultiplier != 1.25 || !group.VisibleOtherGroup ||
		group.Description != "分组说明\n第二行" || group.CreatedAt.IsZero() || !group.CreatedAt.Equal(group.UpdatedAt) {
		t.Fatalf("unexpected created group: %+v", group)
	}
	assertError(t, request(t, api, "POST", "/ui-api/groups", `{"name":" 标准组 ","billing_multiplier":2}`, 409), "group_name_exists")
	other := decode[entity.Group](t, request(t, api, "POST", "/ui-api/groups", `{"name":"其他组","billing_multiplier":2}`, 201))
	if other.VisibleOtherGroup || other.Description != "" {
		t.Fatalf("unexpected defaults: %+v", other)
	}
	assertError(t, request(t, api, "PUT", groupPath,
		`{"name":" 其他组 ","billing_multiplier":5,"visible_other_group":false,"description":"不应保存"}`, 409), "group_name_exists")
	unchanged := decode[entity.Group](t, request(t, api, "GET", groupPath, "", 200))
	if unchanged != group {
		t.Fatalf("failed update changed group: %+v, want %+v", unchanged, group)
	}
	updated := decode[entity.Group](t, request(t, api, "PUT", groupPath,
		`{"name":"  免费组  ","billing_multiplier":0,"visible_other_group":false,"description":""}`, 200))
	if updated.Id != group.Id || updated.Name != "免费组" || updated.BillingMultiplier != 0 || updated.VisibleOtherGroup ||
		updated.Description != "" || !updated.CreatedAt.Equal(group.CreatedAt) || updated.UpdatedAt.Before(group.UpdatedAt) {
		t.Fatalf("unexpected updated group: %+v", updated)
	}
	updated = decode[entity.Group](t, request(t, api, "PUT", groupPath,
		`{"name":"免费组","billing_multiplier":0,"visible_other_group":false,"description":""}`, 200))
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	api, _ = openAPI(t, path)
	persisted := decode[entity.Group](t, request(t, api, "GET", groupPath, "", 200))
	if persisted != updated {
		t.Fatalf("reopened database: %+v, want %+v", persisted, updated)
	}
	listed := decode[[]entity.Group](t, request(t, api, "GET", "/ui-api/groups", "", 200))
	if len(listed) != 2 || listed[0] != other || listed[1] != updated {
		t.Fatalf("unexpected list: %+v", listed)
	}
	deleted := request(t, api, "DELETE", groupPath, "", 204)
	if deleted.Body.Len() != 0 {
		t.Fatalf("delete response must have an empty body: %s", deleted.Body)
	}
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		assertError(t, request(t, api, method, groupPath, `{"name":"不存在","billing_multiplier":1}`, 404), "group_not_found")
	}
	remaining := decode[[]entity.Group](t, request(t, api, "GET", "/ui-api/groups", "", 200))
	if len(remaining) != 1 || remaining[0] != other {
		t.Fatalf("unexpected remaining groups: %+v", remaining)
	}
}

func TestInvalidGroupRequests(t *testing.T) {
	api, _ := openAPI(t, filepath.Join(t.TempDir(), "groups.db"))
	for _, tc := range []struct {
		name, method, path, body, code string
		status                         int
	}{
		{"missing fields", "POST", "/ui-api/groups", `{}`, "validation_error", 400},
		{"null body", "POST", "/ui-api/groups", `null`, "validation_error", 400},
		{"blank name", "POST", "/ui-api/groups", `{"name":"  ","billing_multiplier":1}`, "validation_error", 400},
		{"long name", "POST", "/ui-api/groups", `{"name":"` + strings.Repeat("组", 101) + `","billing_multiplier":1}`, "validation_error", 400},
		{"missing multiplier", "POST", "/ui-api/groups", `{"name":"标准组"}`, "validation_error", 400},
		{"null multiplier", "POST", "/ui-api/groups", `{"name":"标准组","billing_multiplier":null}`, "validation_error", 400},
		{"negative multiplier", "POST", "/ui-api/groups", `{"name":"标准组","billing_multiplier":-0.01}`, "validation_error", 400},
		{"overflow multiplier", "POST", "/ui-api/groups", `{"name":"标准组","billing_multiplier":1e309}`, "invalid_json", 400},
		{"string multiplier", "POST", "/ui-api/groups", `{"name":"标准组","billing_multiplier":"1"}`, "invalid_json", 400},
		{"wrong visibility type", "POST", "/ui-api/groups", `{"name":"标准组","billing_multiplier":1,"visible_other_group":"true"}`, "invalid_json", 400},
		{"long description", "POST", "/ui-api/groups", `{"name":"标准组","billing_multiplier":1,"description":"` + strings.Repeat("说", 1001) + `"}`, "validation_error", 400},
		{"unknown field", "POST", "/ui-api/groups", `{"name":"标准组","billing_multiplier":1,"unknown":true}`, "invalid_json", 400},
		{"malformed JSON", "POST", "/ui-api/groups", `{`, "invalid_json", 400},
		{"multiple values", "POST", "/ui-api/groups", `{"name":"标准组","billing_multiplier":1} {}`, "invalid_json", 400},
		{"oversized body", "POST", "/ui-api/groups", `{"name":"标准组","billing_multiplier":1,"description":"` + strings.Repeat("a", 1<<20) + `"}`, "body_too_large", 413},
		{"invalid update", "PUT", "/ui-api/groups/1", `{"name":"标准组","billing_multiplier":-1}`, "validation_error", 400},
		{"zero Id", "GET", "/ui-api/groups/0", "", "validation_error", 400},
		{"negative Id", "DELETE", "/ui-api/groups/-1", "", "validation_error", 400},
		{"invalid Id", "PUT", "/ui-api/groups/abc", `{}`, "validation_error", 400},
		{"overflow Id", "GET", "/ui-api/groups/9223372036854775808", "", "validation_error", 400},
		{"unsupported method", "PATCH", "/ui-api/groups/1", "", "method_not_allowed", 405},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertError(t, request(t, api, tc.method, tc.path, tc.body, tc.status), tc.code)
		})
	}
	t.Run("content type", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/ui-api/groups", strings.NewReader(`{"name":"标准组","billing_multiplier":1}`))
		req.Header.Set("Content-Type", "text/plain")
		response := httptest.NewRecorder()
		api.ServeHTTP(response, req)
		if response.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("status = %d, want 415", response.Code)
		}
		assertError(t, response, "unsupported_media_type")
	})
	groups := decode[[]entity.Group](t, request(t, api, "GET", "/ui-api/groups", "", 200))
	if len(groups) != 0 {
		t.Fatalf("invalid requests created groups: %+v", groups)
	}
}
