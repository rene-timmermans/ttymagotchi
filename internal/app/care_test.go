package app

import (
	"strings"
	"testing"
	"time"

	"github.com/ezeoleaf/termagotchi/internal/config"
	"github.com/gdamore/tcell/v2"
)

func TestHungerSpeedAnchorsAndProgress(t *testing.T) {
	for speed, want := range map[int]time.Duration{0: 7 * 24 * time.Hour, 50: 8 * time.Hour, 100: 5 * time.Minute} {
		if got := hungerDuration(speed); got != want {
			t.Fatalf("speed %d: %v != %v", speed, got, want)
		}
		a := testApp(t)
		a.updateInterval = hungerDuration(speed) / 100
		a.currentTamagotchi.Hunger = 0
		a.advanceTime(want / 2)
		if a.currentTamagotchi.Hunger != 50 {
			t.Fatalf("speed %d half elapsed hunger: %d", speed, a.currentTamagotchi.Hunger)
		}
		a.advanceTime(want / 2)
		if a.currentTamagotchi.Hunger != 100 {
			t.Fatalf("speed %d full elapsed hunger: %d", speed, a.currentTamagotchi.Hunger)
		}
	}
	for speed := 1; speed <= 100; speed++ {
		if hungerDuration(speed) >= hungerDuration(speed-1) {
			t.Fatalf("nonmonotonic at %d", speed)
		}
	}
}

func TestOverfeedingStarvationAndDeath(t *testing.T) {
	a := testApp(t)
	pet := a.currentTamagotchi
	pet.Hunger = 0
	a.startFeeding(0)
	a.tickActivity(a.activity.started.Add(4 * time.Second))
	if pet.Health != 90 || !strings.Contains(a.activity.message, "Overfed") {
		t.Fatal("overfeeding did not hurt")
	}
	a.closeActivity()
	pet.Hunger = 20
	pet.Health = 100
	a.feedTamagotchi(0)
	if pet.Health != 100 || pet.Hunger != 0 {
		t.Fatal("exact meal caused damage")
	}
	pet.Hunger = 10
	a.feedTamagotchi(0)
	if pet.Health != 95 {
		t.Fatal("excess nutrition did not cause damage")
	}
	pet.Hunger = 89
	pet.Health = 3
	a.advanceTime(a.updateInterval)
	if pet.Health != 1 || !pet.IsAlive {
		t.Fatal("starvation did not damage health")
	}
	a.advanceTime(a.updateInterval)
	if pet.Health != 0 || pet.IsAlive {
		t.Fatal("zero health is not death")
	}
	before := *pet
	a.feedTamagotchi(0)
	a.putTamagotchiToSleep(0)
	a.startCare("wash", 0)
	a.playWithTamagotchi(0)
	if *pet != before || a.activity != nil {
		t.Fatal("dead pet revived")
	}
	deaths := 0
	for _, e := range a.gameEvents {
		if e.Type == "DEATH" {
			deaths++
		}
	}
	if deaths != 1 {
		t.Fatalf("death events: %d", deaths)
	}
	// Feeding can also kill, immediately, before the next time tick.
	b := testApp(t)
	b.currentTamagotchi.Hunger = 0
	b.currentTamagotchi.Health = 1
	b.startFeeding(0)
	b.finishFeeding()
	if b.currentTamagotchi.IsAlive || !strings.Contains(b.activity.message, "passed away") {
		t.Fatal("lethal overfeed did not kill")
	}
}

func TestHygieneDamageGradient(t *testing.T) {
	for value, want := range map[int]int{100: 0, 21: 0, 20: 1, 16: 1, 15: 2, 10: 3, 5: 4, 0: 5} {
		if got := hygieneDamage(value); got != want {
			t.Fatalf("hygiene %d damage=%d want=%d", value, got, want)
		}
		a := testApp(t)
		pet := a.currentTamagotchi
		pet.Hunger = 0
		pet.Cleanliness = value
		pet.Diaper = 100
		a.advanceTime(a.updateInterval)
		if pet.Health != 100-want {
			t.Fatalf("hygiene %d health=%d", value, pet.Health)
		}
	}
	a := testApp(t)
	pet := a.currentTamagotchi
	pet.Cleanliness = 0
	pet.Diaper = 0
	pet.Hunger = 0
	pet.Health = 10
	a.advanceTime(a.updateInterval)
	if pet.Health != 0 || pet.IsAlive {
		t.Fatal("combined neglect did not kill")
	}
}

func TestCareAnimationAndSingleCommit(t *testing.T) {
	cases := []struct {
		kind    string
		index   int
		visible string
	}{{"wash", 0, "Washing"}, {"change", 0, "diaper"}, {"sleep", 0, "Sleeping"}, {"sleep", 1, "Sleeping"}, {"sleep", 2, "Sleeping"}, {"sleep", 3, "Sleeping"}}
	for _, tc := range cases {
		for _, skip := range []bool{false, true} {
			a := testApp(t)
			pet := a.currentTamagotchi
			pet.Cleanliness = 0
			pet.Diaper = 0
			pet.Energy = 0
			pet.Health = 50
			before := *pet
			a.startCare(tc.kind, tc.index)
			g := a.activity
			a.tickActivity(a.activity.started.Add(time.Second))
			if *pet != before {
				t.Fatal("care applied before animation")
			}
			if !strings.Contains(a.activity.view.GetText(true), tc.visible) {
				t.Fatal("missing animation")
			}
			if skip {
				key(a, tcell.KeyEscape, 0)
			} else {
				a.tickActivity(a.activity.started.Add(5 * time.Second))
			}
			after := *pet
			a.tickActivity(time.Now().Add(time.Hour))
			if *pet != after || g.phase != "result" || len(a.gameEvents) != 1 {
				t.Fatal("duplicate care")
			}
			switch tc.kind {
			case "wash":
				if pet.Cleanliness != 100 || pet.Diaper != 0 || pet.Health != 50 || pet.LastWash.IsZero() {
					t.Fatal("wrong bath effect")
				}
			case "change":
				if pet.Cleanliness != 0 || pet.Diaper != 100 || pet.LastChange.IsZero() {
					t.Fatal("wrong change effect")
				}
			case "sleep":
				if pet.Energy != sleepOptions[tc.index].EnergyGain || pet.Health != min(100, 50+sleepOptions[tc.index].HealthGain) {
					t.Fatal("wrong sleep effect")
				}
			}
			if a.activity != nil {
				key(a, tcell.KeyEnter, 0)
			}
			if a.activity != nil {
				t.Fatal("did not return")
			}
		}
	}
}

func TestSleepEyesClosedForAllStages(t *testing.T) {
	for _, stage := range []string{"egg", "baby", "child", "teen", "adult"} {
		pet := *newDefaultTamagotchi("Test")
		pet.Stage = stage
		pet.Happiness = 100
		sprite := strings.Join(actionPose(pet, "sleep", 0), "\n")
		if strings.Contains(sprite, "^") || strings.Contains(sprite, "o o") || strings.Contains(sprite, "o   o") {
			t.Fatalf("eyes open for %s: %s", stage, sprite)
		}
	}
}

func TestCareStopsDamageAndOfflineUsesSpeed(t *testing.T) {
	a := testApp(t)
	pet := a.currentTamagotchi
	pet.Hunger = 0
	pet.Cleanliness = 0
	pet.Diaper = 0
	a.startCare("wash", 0)
	a.finishCare()
	a.closeActivity()
	a.startCare("change", 0)
	a.finishCare()
	a.closeActivity()
	health := pet.Health
	a.advanceTime(a.updateInterval)
	if pet.Health != health {
		t.Fatal("healthy hygiene still caused damage")
	}
	a.Config.App.LastLogin = time.Now().Add(-100 * a.updateInterval)
	a.Config.App.CurrentLogin = a.Config.App.LastLogin.Add(100 * a.updateInterval)
	a.applyOfflineProgress()
	if pet.Hunger != 100 || pet.Health >= health {
		t.Fatal("offline starvation did not apply")
	}
}

func TestSpeedPersistenceAndFraction(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if _, err := config.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	a := testApp(t)
	a.timeAccumulator = a.updateInterval / 2
	a.setSpeed(100)
	if a.updateInterval != 3*time.Second || a.timeAccumulator != 1500*time.Millisecond {
		t.Fatal("lost fractional progress")
	}
	a.setSpeed(0)
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.App.SpeedPercent != 0 || cfg.Tamagotchi.Cleanliness != 100 {
		t.Fatal("not persisted")
	}
	b := &App{Config: cfg, updateInterval: hungerDuration(cfg.App.SpeedPercent) / 100}
	cfg.App.CurrentLogin = cfg.App.LastLogin // Round trip without elapsed time.
	b.initializeStateFromConfig()
	if b.timeAccumulator != b.updateInterval/2 || b.currentTamagotchi.Cleanliness != 100 {
		t.Fatal("roundtrip lost care or fractional state")
	}
}
