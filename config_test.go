package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)      // Linux
	t.Setenv("AppData", tmpDir)               // Windows
	t.Setenv("HOME", tmpDir)                  // macOS fallback

	cfg := Config{Host: "user@myserver", Dir: "/uploads"}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	loaded, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if loaded.Host != "user@myserver" {
		t.Fatalf("expected host 'user@myserver', got %q", loaded.Host)
	}
	if loaded.Dir != "/uploads" {
		t.Fatalf("expected dir '/uploads', got %q", loaded.Dir)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("AppData", tmpDir)
	t.Setenv("HOME", tmpDir)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.Host != "" {
		t.Fatalf("expected empty host, got %q", cfg.Host)
	}
}

func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("AppData", tmpDir)
	t.Setenv("HOME", tmpDir)

	path, err := configPath()
	if err != nil {
		t.Fatalf("configPath: %v", err)
	}
	if filepath.Base(path) != "config.json" {
		t.Fatalf("expected config.json, got %q", filepath.Base(path))
	}
}

func TestSaveConfig_CreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("AppData", tmpDir)
	t.Setenv("HOME", tmpDir)

	cfg := Config{Host: "pi"}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	path, _ := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}
