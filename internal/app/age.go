package app

import (
	"github.com/rene-timmermans/ttymagotchi/internal/config"
	"time"
)

// Three complete hunger periods make one game day. Aging, diaper schedules
// and needs therefore share one clock, including its saved fractional tick.
const ageTicksPerDay = int64(diaperDayTicks)
const maxAgeTicks = int64((1<<63 - 1) / defaultUpdateInterval)

func initialAgeTicks(cfg config.TamagotchiConfig, lastSeen time.Time) int64 {
	if cfg.AgeTicks != nil {
		if *cfg.AgeTicks < 0 {
			return 0
		}
		if *cfg.AgeTicks > maxAgeTicks {
			return maxAgeTicks
		}
		return *cfg.AgeTicks
	}
	// Old saves used calendar days. Preserve their age and stage, importing
	// the fraction of their last saved day before applying offline time once.
	days := max(max(0, cfg.Age), growthDays[growthIndex(cfg.Stage)])
	if int64(days) > maxAgeTicks/ageTicksPerDay {
		return maxAgeTicks
	}
	ticks := int64(days) * ageTicksPerDay
	if !cfg.Created.IsZero() && !lastSeen.Before(cfg.Created) {
		fraction := lastSeen.Sub(cfg.Created) % (24 * time.Hour)
		ticks += int64(fraction / defaultUpdateInterval)
	}
	if ticks > maxAgeTicks {
		return maxAgeTicks
	}
	return ticks
}

func gameDayDuration(speed int) time.Duration { return 3 * hungerDuration(speed) }

func (a *App) ageDuration() time.Duration {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	if a.currentTamagotchi == nil {
		return 0
	}
	elapsed := time.Duration(a.currentTamagotchi.AgeTicks) * defaultUpdateInterval
	if a.updateInterval > 0 && a.currentTamagotchi.IsAlive {
		part := time.Duration(float64(a.timeAccumulator) / float64(a.updateInterval) * float64(defaultUpdateInterval))
		if part > 0 && elapsed <= time.Duration(1<<63-1)-part {
			elapsed += part
		}
	}
	return elapsed
}
