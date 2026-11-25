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
	if cfg.App.SpeedPercent != 50 || cfg.Tamagotchi.Cleanliness != 100 || cfg.Tamagotchi.Diaper != 100 {
		t.Fatal("wrong new defaults")
	}
	path := filepath.Join(cfg.App.SaveDirectory, "config.yml")
	legacy := []byte("app:\n  last_login: 2026-11-25T10:00:00Z\ntamagotchi:\n  name: Original\n  hunger: 25\n  health: 80\n  is_alive: true\n")
	if err := os.WriteFile(path, legacy, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.App.SpeedPercent != 50 || cfg.Tamagotchi.Cleanliness != 100 || cfg.Tamagotchi.Diaper != 100 || cfg.Tamagotchi.Name != "Original" || cfg.Tamagotchi.Health != 80 {
		t.Fatal("legacy save damaged")
	}
	cfg.App.SpeedPercent = 0
	cfg.Tamagotchi.Cleanliness = 0
	cfg.Tamagotchi.Diaper = 0
	cfg.Tamagotchi.Health = 0
	cfg.Tamagotchi.IsAlive = true
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.App.SpeedPercent != 0 || cfg.Tamagotchi.Cleanliness != 0 || cfg.Tamagotchi.Diaper != 0 || cfg.Tamagotchi.IsAlive {
		t.Fatal("explicit zeros/death lost")
	}
}
