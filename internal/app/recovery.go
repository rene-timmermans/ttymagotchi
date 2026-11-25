package app

// Health itself is excluded: a sick pet can recover if its needs are met.
// Hunger is inverted: 60 hunger means 40% of its need for food is still met.
func needsMet(t Tamagotchi) bool {
	return t.IsAlive && t.Hunger <= 60 && t.Happiness >= 40 && t.Energy >= 40 && t.Cleanliness >= 40 && ageCareLevel(t) >= 40
}

func (a *App) resetRecoveryIfNeededLocked() {
	if t := a.currentTamagotchi; t != nil && !needsMet(*t) {
		t.HealthyTicks = 0
	}
}

func (a *App) applyRecoveryLocked() {
	t := a.currentTamagotchi
	if t == nil {
		return
	}
	if !needsMet(*t) || t.Health >= 100 {
		t.HealthyTicks = 0
		return
	}
	t.HealthyTicks++
	// Five consecutive periods of good care: 24 minutes at normal speed.
	if t.HealthyTicks >= 5 {
		t.Health = min(100, t.Health+1)
		t.HealthyTicks = 0
	}
}

func (a *App) recoveryStatus(t Tamagotchi) string {
	if !t.IsAlive {
		return "Recovery: unavailable — your pet has passed away"
	}
	if t.Health >= 100 {
		return "Recovery: fully healthy"
	}
	if !needsMet(t) {
		return "Recovery: keep needs at least 40% (Hunger at most 60%)"
	}
	a.stateMu.RLock()
	interval := a.updateInterval
	a.stateMu.RUnlock()
	return "Recovery: +1 health per " + formatDuration(5*interval) + " of sustained good care"
}
