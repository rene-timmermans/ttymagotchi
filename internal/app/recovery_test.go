package app

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rene-timmermans/ttymagotchi/internal/config"
	"github.com/rivo/tview"
)

func TestRecoveryNeedsAndBoundaries(t *testing.T) {
	base := *newDefaultTamagotchi("Test")
	base.Health = 1
	base.Hunger = 60
	base.Energy = 40
	base.Happiness = 40
	base.Cleanliness = 40
	base.Grooming = 40
	base.Stage = "adult"
	if !needsMet(base) {
		t.Fatal("exact 40% needs should permit recovery, even at health 1")
	}
	for _, change := range []func(*Tamagotchi){func(p *Tamagotchi) { p.Hunger = 61 }, func(p *Tamagotchi) { p.Energy = 39 }, func(p *Tamagotchi) { p.Happiness = 39 }, func(p *Tamagotchi) { p.Cleanliness = 39 }, func(p *Tamagotchi) { p.Grooming = 39 }, func(p *Tamagotchi) { p.IsAlive = false }} {
		pet := base
		change(&pet)
		if needsMet(pet) {
			t.Fatal("unmet need allowed recovery")
		}
	}
}

func TestSustainedRecoveryAtAllSpeeds(t *testing.T) {
	for _, speed := range []int{0, 50, 100} {
		a := testApp(t)
		p := a.currentTamagotchi
		p.Health = 30
		p.Hunger = 0
		a.updateInterval = hungerDuration(speed) / 100
		a.advanceTime(4 * a.updateInterval)
		if p.Health != 30 || p.HealthyTicks != 4 {
			t.Fatal("recovery too soon")
		}
		a.advanceTime(a.updateInterval)
		if p.Health != 31 || p.HealthyTicks != 0 {
			t.Fatal("missing first recovery")
		}
		a.advanceTime(5 * a.updateInterval)
		if p.Health != 32 {
			t.Fatal("missing continued recovery")
		}
	}
}

func TestRecoveryResetsOnNeglectAndOverfeeding(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	p.Hunger = 0
	p.Health = 50
	a.advanceTime(4 * a.updateInterval)
	p.Diaper = 39
	a.advanceTime(a.updateInterval)
	if p.Health != 50 || p.HealthyTicks != 0 {
		t.Fatal("neglect did not reset recovery")
	}
	p.Diaper = 100
	a.advanceTime(4 * a.updateInterval)
	if p.Health != 50 {
		t.Fatal("did not require a fresh recovery period")
	}
	p.Hunger = 0
	a.feedTamagotchi(0)
	if p.HealthyTicks != 0 || p.Health != 40 {
		t.Fatal("overfeeding retained recovery credit")
	}
	p.Energy = 50
	p.HealthyTicks = 4
	a.rewardGame(0, 0, 3)
	if p.HealthyTicks != 0 {
		t.Fatal("activity lowering energy did not reset recovery")
	}
}

func TestRecoveryCapsAndCannotRevive(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	p.Health = 99
	p.Hunger = 0
	a.advanceTime(10 * a.updateInterval)
	if p.Health != 100 || p.HealthyTicks != 0 {
		t.Fatal("recovery exceeded maximum")
	}
	p.Health = 0
	p.IsAlive = false
	p.HealthyTicks = 4
	a.advanceTime(10 * a.updateInterval)
	if p.Health != 0 || p.IsAlive {
		t.Fatal("recovery revived a dead pet")
	}
}

func TestRecoveryPersistsAndRunsOffline(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	a := testApp(t)
	a.Config = cfg
	a.currentTamagotchi.Health = 50
	a.currentTamagotchi.Hunger = 0
	a.advanceTime(4 * a.updateInterval)
	a.updateConfigFromState()
	cfg.App.LastLogin = time.Now()
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	loaded.App.CurrentLogin = loaded.App.LastLogin.Add(a.updateInterval)
	b := testApp(t)
	b.Config = loaded
	b.initializeStateFromConfig()
	if b.currentTamagotchi.Health != 51 || b.currentTamagotchi.HealthyTicks != 0 {
		t.Fatal("offline/reloaded recovery lost progress")
	}
}

func assertStatus(t *testing.T, a *App) {
	t.Helper()
	name, _ := a.TLayout.GetItem(0).(*tview.Pages).GetFrontPage()
	if name != statusSection || a.activity != nil || a.modal != nil {
		t.Fatalf("did not return to Status: %q", name)
	}
	if a.TApp.GetFocus() != a.viewsList["status"] {
		t.Fatal("status does not have keyboard focus")
	}
}

func TestEscapeAndCtrlSFromEveryPageAndRestartDialog(t *testing.T) {
	a := NewApp(&config.Config{App: config.AppConfig{SpeedPercent: 50}})
	defer close(a.done)
	capture := a.TApp.GetInputCapture()
	for _, key := range []tcell.Key{tcell.KeyEscape, tcell.KeyCtrlS} {
		for _, section := range []string{statusSection, feedSection, playSection, sleepSection, hygieneSection, optionsSection, eventsSection, helpSection} {
			a.goToSection(section, nil)
			if event := capture(tcell.NewEventKey(key, 0, tcell.ModNone)); event != nil {
				t.Fatal("navigation key leaked")
			}
			assertStatus(t, a)
		}
		before, _ := a.tamagotchiSnapshot()
		a.showRestartModal()
		capture(tcell.NewEventKey(key, 0, tcell.ModNone))
		assertStatus(t, a)
		after, _ := a.tamagotchiSnapshot()
		if before != after {
			t.Fatal("escape restarted pet")
		}
	}
}

func TestEscapeAndCtrlSFromActivities(t *testing.T) {
	for _, keyCode := range []tcell.Key{tcell.KeyEscape, tcell.KeyCtrlS} {
		for _, kind := range []string{"game", "food", "sleep", "wash", "brush", "diaper"} {
			a := testApp(t)
			if kind == "brush" {
				a.currentTamagotchi.Stage = "adult"
			}
			a.goToSection(playSection, nil)
			if kind == "game" {
				a.playWithTamagotchi(0)
			} else if kind == "food" {
				a.startFeeding(0)
			} else {
				a.startCare(kind, 0)
			}
			key(a, keyCode, 0)
			assertStatus(t, a)
			want := 1
			if kind == "game" || kind == "sleep" {
				want = 0
			}
			if len(a.gameEvents) != want {
				t.Fatal("care duplicated or cancelled game rewarded")
			}
			// Pressing the navigation key again must not repeat care.
			a.returnToStatus()
			if len(a.gameEvents) != want {
				t.Fatal("duplicate care on second Escape")
			}
		}
	}
}
