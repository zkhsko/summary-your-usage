package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTP     HTTP     `json:"http"`
	Database Database `json:"database"`
}

type HTTP struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Database struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config %q: %w", path, err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return Config{}, fmt.Errorf("config %q must contain a single JSON object", path)
	}
	if strings.TrimSpace(cfg.HTTP.Host) == "" {
		return Config{}, fmt.Errorf("http.host is required")
	}
	if cfg.HTTP.Port < 1 || cfg.HTTP.Port > 65535 {
		return Config{}, fmt.Errorf("http.port must be between 1 and 65535")
	}
	if cfg.Database.Driver != "sqlite" && cfg.Database.Driver != "mysql" {
		return Config{}, fmt.Errorf("database.driver must be sqlite or mysql")
	}
	if strings.TrimSpace(cfg.Database.DSN) == "" {
		return Config{}, fmt.Errorf("database.dsn is required")
	}
	return cfg, nil
}

func (c Config) HTTPAddr() string {
	return net.JoinHostPort(c.HTTP.Host, strconv.Itoa(c.HTTP.Port))
}
