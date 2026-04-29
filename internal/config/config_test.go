package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	content := `
[server]
address = ":9090"

[[shares]]
name = "docs"
path = "` + dir + `"
username = "user"
password = "pass"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Address != ":9090" {
		t.Errorf("address = %q, want %q", cfg.Server.Address, ":9090")
	}
	if len(cfg.Shares) != 1 {
		t.Fatalf("shares count = %d, want 1", len(cfg.Shares))
	}
	if cfg.Shares[0].Name != "docs" {
		t.Errorf("share name = %q, want %q", cfg.Shares[0].Name, "docs")
	}
	if cfg.Shares[0].Username != "user" {
		t.Errorf("username = %q, want %q", cfg.Shares[0].Username, "user")
	}
}

func TestLoadDefaultAddress(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	content := `
[[shares]]
name = "docs"
path = "` + dir + `"
username = "user"
password = "pass"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Address != "127.0.0.1:43621" {
		t.Errorf("address = %q, want %q", cfg.Server.Address, "127.0.0.1:43621")
	}
}

func TestLoadNoShares(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	content := `
[server]
address = ":9090"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for no shares")
	}
}

func TestLoadDuplicateShareName(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	content := `
[[shares]]
name = "docs"
path = "` + dir + `"
username = "user"
password = "pass"

[[shares]]
name = "docs"
path = "` + dir + `"
username = "user2"
password = "pass2"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for duplicate share name")
	}
}

func TestLoadMissingCredentials(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	content := `
[[shares]]
name = "docs"
path = "` + dir + `"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Shares[0].Username != "" || cfg.Shares[0].Password != "" {
		t.Errorf("expected empty credentials, got user=%q pass=%q",
			cfg.Shares[0].Username, cfg.Shares[0].Password)
	}
}

func TestLoadNonExistentPath(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	content := `
[[shares]]
name = "docs"
path = "/nonexistent/path/that/does/not/exist"
username = "user"
password = "pass"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.toml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadTildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(home, ".config", "local-webdav", "config.toml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Skip("default config not found, skipping tilde expansion integration test")
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, s := range cfg.Shares {
		if strings.Contains(s.Path, "~") {
			t.Errorf("share %q path still contains ~: %q", s.Name, s.Path)
		}
		if !filepath.IsAbs(s.Path) {
			t.Errorf("share %q path is not absolute: %q", s.Name, s.Path)
		}
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/Documents", filepath.Join(home, "Documents")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~", "~"},
	}

	for _, tt := range tests {
		got, err := expandHome(tt.input)
		if err != nil {
			t.Errorf("expandHome(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDefaultPath(t *testing.T) {
	p, err := DefaultPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == "" {
		t.Fatal("default path is empty")
	}
	if filepath.Ext(p) != ".toml" {
		t.Errorf("default path %q does not end with .toml", p)
	}
}

func TestEnsureConfig_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "sub", "config.toml")

	created, err := EnsureConfig(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected created=true")
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("config file is empty")
	}
}

func TestEnsureConfig_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")

	original := `[server]
address = ":9999"
`
	if err := os.WriteFile(cfgPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	created, err := EnsureConfig(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Fatal("expected created=false for existing file")
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Errorf("EnsureConfig overwrote existing file")
	}
}
