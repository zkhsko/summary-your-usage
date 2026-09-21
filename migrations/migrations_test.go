package migrations_test

import (
	"path/filepath"
	"testing"

	"summary-your-usage/migrations"
	"summary-your-usage/pkg/database"
)

func TestSQLiteMigrationLifecycle(t *testing.T) {
	db, err := database.Open(t.Context(), "sqlite", filepath.Join(t.TempDir(), "migrations.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	provider, err := migrations.New(db, "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(t.Context(), 1); err != nil {
		t.Fatalf("up users: %v", err)
	}
	if _, err := db.ExecContext(t.Context(), "INSERT INTO users(name, email, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", "Alice", "alice@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(t.Context()); err != nil {
		t.Fatalf("up groups: %v", err)
	}
	if _, err := db.ExecContext(t.Context(), "INSERT INTO `groups`(name, created_at, updated_at) VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", "标准组"); err != nil {
		t.Fatal(err)
	}
	results, err := provider.Up(t.Context())
	if err != nil || len(results) != 0 {
		t.Fatalf("repeated up: results = %v, err = %v", results, err)
	}
	var count int
	if err := db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM users").Scan(&count); err != nil || count != 1 {
		t.Fatalf("repeated up must preserve data: count = %d, err = %v", count, err)
	}
	var multiplier float64
	var visible bool
	var description string
	if err := db.QueryRowContext(t.Context(), "SELECT billing_multiplier, visible_other_group, description FROM `groups` WHERE name = ?", "标准组").Scan(&multiplier, &visible, &description); err != nil || multiplier != 1 || visible || description != "" {
		t.Fatalf("group defaults: multiplier = %v, visible = %v, description = %q, err = %v", multiplier, visible, description, err)
	}
	if _, err := provider.Down(t.Context()); err != nil {
		t.Fatalf("down groups: %v", err)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'groups'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("down must remove groups: count = %d, err = %v", count, err)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM users").Scan(&count); err != nil || count != 1 {
		t.Fatalf("group rollback must preserve users: count = %d, err = %v", count, err)
	}
	if version, err := provider.GetDBVersion(t.Context()); err != nil || version != 1 {
		t.Fatalf("group rollback version = %d, err = %v", version, err)
	}
	if _, err := provider.Down(t.Context()); err != nil {
		t.Fatalf("down: %v", err)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'users'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("down must remove users: count = %d, err = %v", count, err)
	}
	if version, err := provider.GetDBVersion(t.Context()); err != nil || version != 0 {
		t.Fatalf("down version = %d, err = %v", version, err)
	}
	if _, err := provider.Up(t.Context()); err != nil {
		t.Fatalf("up after down: %v", err)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM users").Scan(&count); err != nil || count != 0 {
		t.Fatalf("up must recreate an empty table: count = %d, err = %v", count, err)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM `groups`").Scan(&count); err != nil || count != 0 {
		t.Fatalf("up must recreate an empty groups table: count = %d, err = %v", count, err)
	}
	if version, err := provider.GetDBVersion(t.Context()); err != nil || version != 2 {
		t.Fatalf("up version = %d, err = %v", version, err)
	}
}
