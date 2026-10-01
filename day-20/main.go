package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l LogLevel) String() string {
	return [...]string{"debug", "info", "warn", "error"}[l]
}

func (l LogLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

func (l *LogLevel) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "debug":
		*l = LevelDebug
	case "info":
		*l = LevelInfo
	case "warn":
		*l = LevelWarn
	case "error":
		*l = LevelError
	default:
		return fmt.Errorf("unknown log level %q", s)
	}
	return nil
}

type DatabaseConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Name string `json:"name"`
}

type Config struct {
	Server   string         `json:"server"`
	Port     int            `json:"port"`
	Database DatabaseConfig `json:"database"`
	LogLevel LogLevel       `json:"log_level"`
	Debug    bool           `json:"debug"`
}

func defaultConfig() Config {
	return Config{
		Server:   "localhost",
		Port:     8080,
		Database: DatabaseConfig{Host: "localhost", Port: 5432, Name: "mydb"},
		LogLevel: LevelInfo,
	}
}

func loadFromJSON(data []byte) (Config, error) {
	cfg := defaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal: %w", err)
	}
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		var l LogLevel
		if err := json.Unmarshal([]byte(`"`+v+`"`), &l); err == nil {
			cfg.LogLevel = l
		}
	}
	if v := os.Getenv("DEBUG"); v == "true" || v == "1" {
		cfg.Debug = true
	}
}

func exportCSV(cfg Config) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"key", "value"})
	w.Write([]string{"server", cfg.Server})
	w.Write([]string{"port", strconv.Itoa(cfg.Port)})
	w.Write([]string{"db_host", cfg.Database.Host})
	w.Write([]string{"db_port", strconv.Itoa(cfg.Database.Port)})
	w.Write([]string{"db_name", cfg.Database.Name})
	w.Write([]string{"log_level", cfg.LogLevel.String()})
	w.Write([]string{"debug", strconv.FormatBool(cfg.Debug)})
	w.Flush()
	return buf.String()
}

func main() {
	jsonData := []byte(`{
		"server": "api.example.com",
		"port": 443,
		"database": {"host": "db.example.com", "port": 5432, "name": "prod"},
		"log_level": "warn",
		"debug": false
	}`)

	cfg, err := loadFromJSON(jsonData)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	applyEnvOverrides(&cfg)

	out, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println("=== Loaded Config (JSON) ===")
	fmt.Println(string(out))

	fmt.Println("\n=== CSV Export ===")
	fmt.Print(exportCSV(cfg))
}
