package app

import (
	"testing"
	"time"

	"github.com/ezeoleaf/termagotchi/internal/config"
)

func fixedDiaperSchedule(p *Tamagotchi) {
	p.DiaperEventCount = 5
	p.DiaperEvents = [5]int{20, 80, 140, 200, 260}
	p.DiaperDayTicks = 0
	p.DiaperNextEvent = 0
}

func TestDiaperChangesOnlyAtBowelEvents(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	fixedDiaperSchedule(p)
	a.advanceTime(19 * a.updateInterval)
	if p.Diaper != 100 || len(a.gameEvents) != 0 {
		t.Fatal("diaper decayed between bowel events")
	}
	a.advanceTime(a.updateInterval)
	if p.Diaper != 65 || len(a.gameEvents) != 1 || a.gameEvents[0].Type != "DIAPER" {
		t.Fatal("first bowel event did not immediately subtract 35")
	}
	for _, point := range []struct{ at, want int }{{79, 65}, {80, 30}, {139, 30}, {140, 0}, {200, 0}, {260, 0}} {
		for p.DiaperDayTicks < point.at {
			a.advanceDiaperLocked()
		}
		if p.Diaper != point.want {
			t.Fatalf("day position %d: diaper %d, want %d", point.at, p.Diaper, point.want)
		}
	}
	if len(a.gameEvents) != 5 {
		t.Fatal("day did not contain exactly five events")
	}
}

func TestFourOrFiveRandomlyTimedEventsPerDay(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	counts := map[int]bool{}
	schedules := map[[5]int]bool{}
	for day := 0; day < 40; day++ {
		if !validDiaperSchedule(*p) {
			t.Fatal("invalid generated schedule")
		}
		count := p.DiaperEventCount
		counts[count] = true
		schedules[p.DiaperEvents] = true
		events := 0
		for step := 0; step < diaperDayTicks; step++ {
			before := p.DiaperNextEvent
			a.advanceDiaperLocked()
			if p.DiaperNextEvent > before {
				events++
			}
		}
		if events != count || p.DiaperDayTicks != 0 {
			t.Fatalf("day %d had %d events instead of %d", day, events, count)
		}
	}
	if len(counts) != 2 || len(schedules) < 2 {
		t.Fatal("daily count and timing do not vary")
	}
	if diaperDayTicks*defaultUpdateInterval != 24*time.Hour {
		t.Fatal("normal day is not 24 hours")
	}
}

func TestChangingDiaperDoesNotPostponeEvents(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	fixedDiaperSchedule(p)
	p.DiaperDayTicks = 19
	p.Diaper = 30
	schedule, count, next, clock := p.DiaperEvents, p.DiaperEventCount, p.DiaperNextEvent, p.DiaperDayTicks
	a.startCare("diaper", 0)
	a.finishCare()
	a.closeActivity()
	if p.Diaper != 100 || p.DiaperEvents != schedule || p.DiaperEventCount != count || p.DiaperNextEvent != next || p.DiaperDayTicks != clock {
		t.Fatal("changing a diaper changed the bowel schedule")
	}
	a.advanceTime(a.updateInterval)
	if p.Diaper != 65 || p.DiaperNextEvent != 1 {
		t.Fatal("changing a diaper delayed an imminent bowel event")
	}
}

func TestDiaperScheduleSurvivesReloadAndOfflineTime(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	a := testApp(t)
	a.Config = cfg
	p := a.currentTamagotchi
	fixedDiaperSchedule(p)
	p.DiaperDayTicks = 19
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
	pet := b.currentTamagotchi
	if pet.DiaperEvents != p.DiaperEvents || pet.DiaperEventCount != 5 || pet.DiaperDayTicks != 20 || pet.DiaperNextEvent != 1 || pet.Diaper != 65 {
		t.Fatal("reload/offline time rerolled or lost the scheduled event")
	}
	b.advanceTime(b.updateInterval)
	if pet.Diaper != 65 || pet.DiaperNextEvent != 1 {
		t.Fatal("offline event was applied twice")
	}
}

func TestSpeedScalesTimingWithoutRerollingSchedule(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	a := testApp(t)
	a.Config = cfg
	p := a.currentTamagotchi
	fixedDiaperSchedule(p)
	p.DiaperDayTicks = 19
	a.timeAccumulator = a.updateInterval / 2
	a.setSpeed(100)
	if p.DiaperEvents != [5]int{20, 80, 140, 200, 260} || p.DiaperDayTicks != 19 {
		t.Fatal("speed change rerolled the schedule")
	}
	a.advanceTime(1499 * time.Millisecond)
	if p.Diaper != 100 {
		t.Fatal("event happened too early")
	}
	a.advanceTime(time.Millisecond)
	if p.Diaper != 65 {
		t.Fatal("event did not scale with the clock")
	}
}

func TestBowelEventsStopWhenDiapersAreOutgrown(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	fixedDiaperSchedule(p)
	p.Stage = "toddler"
	p.Created = time.Now().Add(-3 * 24 * time.Hour)
	p.DiaperDayTicks = 19
	p.Diaper = 37
	a.advanceTime(a.updateInterval)
	if p.Diaper != 37 || p.DiaperDayTicks != 19 || len(a.gameEvents) != 0 {
		t.Fatal("bowel schedule still affects a pet that outgrew diapers")
	}
}

func TestSoiledDiaperClampsAndAffectsHealth(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	fixedDiaperSchedule(p)
	p.DiaperDayTicks = 19
	p.Diaper = 20
	a.advanceTime(a.updateInterval)
	if p.Diaper != 0 || p.Health != 95 || p.Happiness != 49 || hygieneLevel(*p) != 0 {
		t.Fatal("a full diaper did not apply neglect consequences")
	}
}

func TestLegacyOrInvalidScheduleStartsWithoutSoiling(t *testing.T) {
	p := newDefaultTamagotchi("Existing")
	p.Diaper = 42
	p.DiaperEventCount = 0
	p.DiaperEvents = [5]int{}
	ensureDiaperSchedule(p)
	if !validDiaperSchedule(*p) || p.Diaper != 42 || p.DiaperDayTicks != 0 {
		t.Fatal("migration damaged the diaper or did not create a schedule")
	}
	p.DiaperNextEvent = 99
	ensureDiaperSchedule(p)
	if !validDiaperSchedule(*p) || p.Diaper != 42 {
		t.Fatal("invalid saved schedule was not repaired safely")
	}
}
