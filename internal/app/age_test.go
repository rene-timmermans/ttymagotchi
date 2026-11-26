package app

import (
	"github.com/rene-timmermans/ttymagotchi/internal/config"
	"testing"
	"time"
)

func advanceWithCare(a *App, ticks int) {
	for i := 0; i < ticks; i++ {
		p := a.currentTamagotchi
		p.Hunger = 0
		p.Energy = 100
		p.Happiness = 100
		p.Health = 100
		p.Cleanliness = 100
		p.Diaper = 100
		p.Grooming = 100
		a.advanceTime(a.updateInterval)
	}
}

func TestAgeSpeedMatchesNeedsAndHatchesAtOneGameDay(t *testing.T) {
	for _, speed := range []int{0, 50, 100} {
		a := testApp(t)
		a.updateInterval = hungerDuration(speed) / 100
		advanceWithCare(a, 299)
		if a.currentTamagotchi.Stage != "egg" || a.currentTamagotchi.Age != 0 {
			t.Fatal("hatched too soon")
		}
		advanceWithCare(a, 1)
		if a.currentTamagotchi.Stage != "hatchling" || a.currentTamagotchi.Age != 1 {
			t.Fatalf("speed %d did not grow with the needs clock", speed)
		}
		if a.ageDuration() != 24*time.Hour {
			t.Fatal("one game day must equal 24 simulated hours")
		}
	}
	for speed, want := range map[int]time.Duration{0: 21 * 24 * time.Hour, 50: 24 * time.Hour, 100: 15 * time.Minute} {
		if gameDayDuration(speed) != want {
			t.Fatal("incorrect real duration for a game day")
		}
	}
}

func TestSpeedChangePreservesWholeAndFractionalAge(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	a := testApp(t)
	a.currentTamagotchi.AgeTicks = 299
	a.advanceTime(a.updateInterval / 2)
	before := a.ageDuration()
	a.setSpeed(100)
	if delta := a.ageDuration() - before; delta < -time.Microsecond || delta > time.Microsecond {
		t.Fatal("changing speed changed past age", delta)
	}
	a.advanceTime(1500 * time.Millisecond)
	if a.currentTamagotchi.Stage != "hatchling" {
		t.Fatal("remaining half tick did not use the new speed")
	}
	a.setSpeed(0)
	if a.currentTamagotchi.Age != 1 {
		t.Fatal("slower speed rolled age back")
	}
}

func TestAgePersistsAndOfflineTimeIsAppliedOnce(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	a := testApp(t)
	a.Config = cfg
	a.updateInterval = 3 * time.Second
	cfg.App.SpeedPercent = 100
	a.currentTamagotchi.AgeTicks = 299
	a.currentTamagotchi.Hunger = 0
	a.timeAccumulator = time.Second
	a.updateConfigFromState()
	cfg.App.LastLogin = time.Now()
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	saved, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	saved.App.CurrentLogin = saved.App.LastLogin.Add(2 * time.Second)
	b := testApp(t)
	b.Config = saved
	b.updateInterval = 3 * time.Second
	b.initializeStateFromConfig()
	if b.currentTamagotchi.AgeTicks != 300 || b.currentTamagotchi.Stage != "hatchling" {
		t.Fatal("fractional save/offline growth lost")
	}
	b.applyOfflineProgress()
	if b.currentTamagotchi.AgeTicks != 300 {
		t.Fatal("offline age was counted twice")
	}
	b.currentTamagotchi.Health = 0
	b.currentTamagotchi.IsAlive = false
	b.advanceTime(time.Hour)
	if b.currentTamagotchi.AgeTicks != 300 {
		t.Fatal("dead pet kept aging")
	}
}

func TestLegacyAgeImportKeepsAgeAndNeverUsesNewSpeedRetroactively(t *testing.T) {
	seen := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	cfg := config.TamagotchiConfig{Age: 2, Stage: "baby", Created: seen.Add(-60 * time.Hour)}
	if ticks := initialAgeTicks(cfg, seen); ticks != 750 {
		t.Fatalf("lost two days and half-day progress: %d", ticks)
	}
	cfg.AgeTicks = new(int64)
	if ticks := initialAgeTicks(cfg, seen); ticks != 0 {
		t.Fatal("explicit zero age was treated as a legacy save")
	}
}
