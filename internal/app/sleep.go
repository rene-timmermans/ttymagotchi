package app

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type SleepOption struct {
	Name       string
	Duration   time.Duration
	EnergyGain int
	HealthGain int
	Happiness  int
}

var sleepOptions = []SleepOption{
	{Name: "😴 Short Nap (30 min)", Duration: 30 * time.Minute, EnergyGain: 20, HealthGain: 5, Happiness: 5},
	{Name: "😪 Medium Sleep (2 hours)", Duration: 2 * time.Hour, EnergyGain: 50, HealthGain: 15, Happiness: 10},
	{Name: "😴 Long Sleep (6 hours)", Duration: 6 * time.Hour, EnergyGain: 80, HealthGain: 25, Happiness: 15},
	{Name: "😴 Full Night (8 hours)", Duration: 8 * time.Hour, EnergyGain: 100, HealthGain: 30, Happiness: 20},
}

func (a *App) generateSleepList(listSleep *tview.List) {
	defer preserveSelection(listSleep)()
	listSleep.Clear()

	t, ok := a.tamagotchiSnapshot()
	if !ok {
		listSleep.AddItem(a.tr("No tamagotchi available."), "", 0, nil)
		return
	}

	if !t.IsAlive {
		listSleep.AddItem(a.tr("Your tamagotchi has passed away... 💔"), "", 0, nil)
		listSleep.AddItem(a.tr("Cannot put a dead tamagotchi to sleep"), "", 0, nil)
		return
	}

	listSleep.AddItem(a.tr("=== SLEEP OPTIONS ==="), "", 0, nil)
	listSleep.AddItem("", "", 0, nil) // Empty line

	for i, sleep := range sleepOptions {
		sleepIndex := i // Capture the index for the closure
		listSleep.AddItem(
			fmt.Sprintf(a.tr("%s (Energy: +%d, Health: +%d, Happiness: +%d)"),
				a.tr(sleep.Name), sleep.EnergyGain, sleep.HealthGain, sleep.Happiness),
			"",
			0,
			func() { a.startCare("sleep", sleepIndex) },
		)
	}

	listSleep.AddItem("", "", 0, nil) // Empty line
	listSleep.AddItem(a.tr("=== SLEEP INFO ==="), "", 0, nil)
	listSleep.AddItem(fmt.Sprintf(a.tr("Current Energy: %d/100"), t.Energy), "", 0, nil)
	listSleep.AddItem(fmt.Sprintf(a.tr("Current Health: %d/100"), t.Health), "", 0, nil)
	listSleep.AddItem(fmt.Sprintf(a.tr("Current Happiness: %d/100"), t.Happiness), "", 0, nil)
	listSleep.AddItem(fmt.Sprintf(a.tr("Last Sleep: %s"), t.LastSleep.Format("15:04")), "", 0, nil)

	// Show sleep recommendation
	if t.Energy < 30 {
		listSleep.AddItem("", "", 0, nil) // Empty line
		listSleep.AddItem(a.tr("💡 Recommendation: Your tamagotchi is tired!"), "", 0, nil)
		listSleep.AddItem(a.tr("   Consider a longer sleep to restore energy."), "", 0, nil)
	}
}

func (a *App) putTamagotchiToSleep(sleepIndex int) {
	if sleepIndex < 0 || sleepIndex >= len(sleepOptions) {
		return
	}

	sleep := sleepOptions[sleepIndex]
	now := time.Now()

	a.stateMu.Lock()
	if a.currentTamagotchi == nil || !a.currentTamagotchi.IsAlive || a.currentTamagotchi.Energy > 60 {
		a.stateMu.Unlock()
		return
	}

	a.currentTamagotchi.Energy = min(100, a.currentTamagotchi.Energy+sleep.EnergyGain)
	a.currentTamagotchi.Health = min(100, a.currentTamagotchi.Health+sleep.HealthGain)
	a.currentTamagotchi.Happiness = min(100, a.currentTamagotchi.Happiness+sleep.Happiness)
	a.currentTamagotchi.LastSleep = now
	a.resetRecoveryIfNeededLocked()
	a.stateMu.Unlock()

	a.updateConfigFromState()
	a.addLocalizedEvent("SLEEP", "Slept for %s! Energy +%d, Health +%d", translatedArgument(sleep.Name), sleep.EnergyGain, sleep.HealthGain)
}

func (a *App) sleepPage() (title string, content tview.Primitive) {
	listSleep := a.viewsList["sleep"]
	if listSleep == nil {
		listSleep = getList()
		a.viewsList["sleep"] = listSleep
	}

	a.generateSleepList(listSleep)

	title = sleepSection
	return title, tview.NewFlex().
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(listSleep, 0, 1, true), 0, 1, true)
}
