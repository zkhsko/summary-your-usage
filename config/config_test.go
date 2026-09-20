package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"summary-your-usage/config"
)

const sqliteConfig = `{"http":{"host":"127.0.0.1","port":8080},"database":{"driver":"sqlite","dsn":"file:test.db"}}`

func configFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	for key, value := range map[string]string{
		"HTTP_ADDR": "ignored:1234",
		"DB_DRIVER": "ignored",
		"DB_DSN":    "ignored",
		"UI_DIR":    "/ignored",
	} {
		t.Setenv(key, value)
	}
	for _, tc := range []struct {
		name, content, address string
		want                   config.Config
	}{
		{
			name: "sqlite", content: sqliteConfig, address: "127.0.0.1:8080",
			want: config.Config{HTTP: config.HTTP{Host: "127.0.0.1", Port: 8080}, Database: config.Database{Driver: "sqlite", DSN: "file:test.db"}},
		},
		{
			name: "mysql", content: `{"http":{"host":"localhost","port":65535},"database":{"driver":"mysql","dsn":"app:secret@tcp(localhost:3306)/app?parseTime=true"}}`, address: "localhost:65535",
			want: config.Config{HTTP: config.HTTP{Host: "localhost", Port: 65535}, Database: config.Database{Driver: "mysql", DSN: "app:secret@tcp(localhost:3306)/app?parseTime=true"}},
		},
		{
			name: "ipv6", content: `{"http":{"host":"::1","port":1},"database":{"driver":"sqlite","dsn":":memory:"}}`, address: "[::1]:1",
			want: config.Config{HTTP: config.HTTP{Host: "::1", Port: 1}, Database: config.Database{Driver: "sqlite", DSN: ":memory:"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.Load(configFile(t, tc.content))
			if err != nil {
				t.Fatal(err)
			}
			if cfg != tc.want {
				t.Fatalf("config = %+v, want %+v", cfg, tc.want)
			}
			if got := cfg.HTTPAddr(); got != tc.address {
				t.Fatalf("HTTPAddr = %q, want %q", got, tc.address)
			}
		})
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	for _, tc := range []struct {
		name, content string
	}{
		{"empty file", ""},
		{"malformed JSON", "{"},
		{"null", "null"},
		{"multiple objects", sqliteConfig + " {}"},
		{"trailing garbage", sqliteConfig + " invalid"},
		{"missing http", `{"database":{"driver":"sqlite","dsn":"file:test.db"}}`},
		{"missing database", `{"http":{"host":"127.0.0.1","port":8080}}`},
		{"blank host", strings.Replace(sqliteConfig, `"127.0.0.1"`, `" "`, 1)},
		{"negative port", strings.Replace(sqliteConfig, "8080", "-1", 1)},
		{"zero port", strings.Replace(sqliteConfig, "8080", "0", 1)},
		{"port too high", strings.Replace(sqliteConfig, "8080", "65536", 1)},
		{"wrong port type", strings.Replace(sqliteConfig, "8080", `"8080"`, 1)},
		{"missing driver", strings.Replace(sqliteConfig, `"driver":"sqlite",`, "", 1)},
		{"unsupported driver", strings.Replace(sqliteConfig, `"sqlite"`, `"postgres"`, 1)},
		{"missing dsn", strings.Replace(sqliteConfig, `,"dsn":"file:test.db"`, "", 1)},
		{"blank dsn", strings.Replace(sqliteConfig, `"file:test.db"`, `" "`, 1)},
		{"unknown field", strings.Replace(sqliteConfig, `"host"`, `"hostname"`, 1)},
		{"ui directory is not configurable", strings.Replace(sqliteConfig, `"http":`, `"ui_dir":"/tmp/ui","http":`, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := config.Load(configFile(t, tc.content)); err == nil {
				t.Fatal("expected invalid configuration to fail")
			}
		})
	}
	t.Run("missing file", func(t *testing.T) {
		if _, err := config.Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
			t.Fatal("expected missing configuration file to fail")
		}
	})
}
