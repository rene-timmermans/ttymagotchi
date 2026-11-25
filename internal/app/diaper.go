package app

import "math/rand/v2"

// One hunger period represents 4m48s at normal speed: 300 periods = 24h.
// Schedules use this game clock, so speed changes preserve their progress.
const diaperDayTicks = 300

func validDiaperSchedule(t Tamagotchi) bool {
	if (t.DiaperEventCount != 4 && t.DiaperEventCount != 5) || t.DiaperDayTicks < 0 || t.DiaperDayTicks >= diaperDayTicks || t.DiaperNextEvent < 0 || t.DiaperNextEvent > t.DiaperEventCount {
		return false
	}
	previous := 0
	for i := 0; i < t.DiaperEventCount; i++ {
		at := t.DiaperEvents[i]
		if at <= previous || at >= diaperDayTicks {
			return false
		}
		if (i < t.DiaperNextEvent && at > t.DiaperDayTicks) || (i >= t.DiaperNextEvent && at <= t.DiaperDayTicks) {
			return false
		}
		previous = at
	}
	return true
}

func newDiaperSchedule(t *Tamagotchi) {
	t.DiaperDayTicks = 0
	t.DiaperNextEvent = 0
	t.DiaperEvents = [5]int{}
	t.DiaperEventCount = 4 + rand.IntN(2)
	// Jitter within separate daily windows avoids unrealistic clusters.
	window := diaperDayTicks / t.DiaperEventCount
	for i := 0; i < t.DiaperEventCount; i++ {
		t.DiaperEvents[i] = i*window + window/4 + rand.IntN(window/2)
	}
}

func ensureDiaperSchedule(t *Tamagotchi) {
	if !validDiaperSchedule(*t) {
		newDiaperSchedule(t)
	}
}

// Call from the simulation loop while holding stateMu. Changing a diaper does
// not alter this schedule: fresh diapers cannot postpone the next bowel event.
func (a *App) advanceDiaperLocked() {
	t := a.currentTamagotchi
	if t == nil || !t.IsAlive || !needsDiaper(*t) {
		return
	}
	ensureDiaperSchedule(t)
	t.DiaperDayTicks++
	if t.DiaperNextEvent < t.DiaperEventCount && t.DiaperDayTicks >= t.DiaperEvents[t.DiaperNextEvent] {
		before := t.Diaper
		t.Diaper = max(0, t.Diaper-35)
		t.DiaperNextEvent++
		a.addLocalizedEvent("DIAPER", "Your pet soiled its diaper. Diaper %d → %d.", before, t.Diaper)
	}
	if t.DiaperDayTicks == diaperDayTicks {
		newDiaperSchedule(t)
	}
}
