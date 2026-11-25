package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLegacySaveMigrationAndZeroSpeed(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.App.SpeedPercent != 50 || cfg.Tamagotchi.Cleanliness != 100 || cfg.Tamagotchi.Grooming != 100 {
		t.Fatal("wrong new defaults")
	}
	path := filepath.Join(cfg.App.SaveDirectory, "config.yml")
	legacy := []byte("app:\n  last_login: 2026-11-25T10:00:00Z\ntamagotchi:\n  name: Original\n  hunger: 25\n  health: 80\n  diaper: 0\n  is_alive: true\n")
	if err := os.WriteFile(path, legacy, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.App.SpeedPercent != 50 || cfg.Tamagotchi.Cleanliness != 100 || cfg.Tamagotchi.Grooming != 100 || cfg.Tamagotchi.Name != "Original" || cfg.Tamagotchi.Health != 80 {
		t.Fatal("legacy save damaged")
	}
	cfg.App.SpeedPercent = 0
	cfg.Tamagotchi.Cleanliness = 0
	cfg.Tamagotchi.Grooming = 0
	cfg.Tamagotchi.GroomTicks = 20
	cfg.Tamagotchi.Health = 0
	cfg.Tamagotchi.IsAlive = true
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.App.SpeedPercent != 0 || cfg.Tamagotchi.Cleanliness != 0 || cfg.Tamagotchi.Grooming != 0 || cfg.Tamagotchi.GroomTicks != 20 || cfg.Tamagotchi.IsAlive {
		t.Fatal("explicit zeros/death lost")
	}
}

func TestRenameImportsLegacySaveWithoutOverwritingIt(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	oldDir := filepath.Join(root, "termagotchi")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(oldDir, "config.yml")
	original := []byte("app:\n  speed_percent: 0\n  save_directory: /old/path\ntamagotchi:\n  name: Kept\n  health: 72\n  diaper: 30\n  species: cat\n  is_alive: true\n")
	if err := os.WriteFile(oldPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Tamagotchi.Name != "Kept" || cfg.Tamagotchi.Health != 72 || cfg.Tamagotchi.Diaper != 30 || cfg.App.SpeedPercent != 0 || cfg.App.SaveDirectory != filepath.Join(root, "ttymagotchi") {
		t.Fatal("legacy save not preserved", cfg)
	}
	cfg.Tamagotchi.Name = "New save"
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(oldPath)
	if err != nil || string(old) != string(original) {
		t.Fatal("old save modified")
	}
	cfg, err = LoadConfig()
	if err != nil || cfg.Tamagotchi.Name != "New save" {
		t.Fatal("new save must take precedence")
	}
}
