package app

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

// Each neglected need contributes independently; damage increases toward zero.
func hygieneDamage(value int) int {
	if value > 20 {
		return 0
	}
	return 1 + (20-max(0, value))/5
}

func (a *App) generateHygieneList(list *tview.List) {
	defer preserveSelection(list)()
	list.Clear()
	t, ok := a.tamagotchiSnapshot()
	if !ok {
		list.AddItem("No tamagotchi available.", "", 0, nil)
		return
	}
	if !t.IsAlive {
		list.AddItem("Your tamagotchi has passed away... 💔", "", 0, nil)
		return
	}
	list.AddItem("=== HYGIENE ===", "", 0, nil)
	list.AddItem("Wash — a bubbly bath", "", 0, func() { a.startCare("wash", 0) })
	if needsDiaper(t) {
		list.AddItem("Diaper — change into a fresh diaper", "", 0, func() { a.startCare("diaper", 0) })
	} else {
		list.AddItem("Brushing — care for fur and feathers", "", 0, func() { a.startCare("brush", 0) })
	}
	list.AddItem("", "", 0, nil)
	list.AddItem("Hygiene: "+a.createProgressBar(hygieneLevel(t), 100), "", 0, nil)
	list.AddItem("Cleanliness: "+a.createProgressBar(t.Cleanliness, 100), "", 0, nil)
	list.AddItem(ageCareLabel(t)+": "+a.createProgressBar(ageCareLevel(t), 100), "", 0, nil)
	list.AddItem(fmt.Sprintf("Health: %d/100", t.Health), "", 0, nil)
	list.AddItem("", "", 0, nil)
	list.AddItem("At 20% or below, neglected hygiene damages health.", "", 0, nil)
	list.AddItem("Closer to 0% = more damage. Each active need counts independently.", "", 0, nil)
	if needsDiaper(t) {
		list.AddItem("Diapers are needed through stage 3; Brushing starts at stage 4.", "", 0, nil)
	} else {
		list.AddItem("Brushing declines very slowly: 100 → 0 in ~7 days at 50% speed.", "", 0, nil)
	}
	list.AddItem("Baths restore cleanliness; the second action restores age-specific care.", "", 0, nil)
	careName, careAt := "Last brushing", t.LastBrush
	if needsDiaper(t) {
		careName, careAt = "Last diaper change", t.LastChange
	}
	for _, action := range []struct {
		name string
		at   time.Time
	}{{"Last bath", t.LastWash}, {careName, careAt}} {
		when := "Not yet"
		if !action.at.IsZero() {
			when = action.at.Format("02 Jan 15:04")
		}
		list.AddItem(action.name+": "+when, "", 0, nil)
	}
}

func (a *App) hygienePage() (string, tview.Primitive) {
	list := getList()
	a.viewsList["hygiene"] = list
	a.generateHygieneList(list)
	return hygieneSection, list
}

func isCare(kind string) bool {
	return kind == "sleep" || kind == "wash" || kind == "brush" || kind == "diaper"
}

func (a *App) startCare(kind string, index int) {
	if !isCare(kind) || a.activity != nil || a.modal != nil {
		return
	}
	if kind == "sleep" && (index < 0 || index >= len(sleepOptions)) {
		return
	}
	t, ok := a.tamagotchiSnapshot()
	if !ok || !t.IsAlive || ((kind == "brush" || kind == "diaper") && kind != ageCareKind(t)) {
		return
	}
	a.openActivity(kind, index)
	a.activity.started = time.Now()
	a.activity.phase = "playing"
	a.renderActivity(time.Now())
}

func (a *App) finishCare() {
	g := a.activity
	if g.phase == "result" {
		return
	}
	g.phase = "result"
	before, _ := a.tamagotchiSnapshot()
	if !before.IsAlive {
		g.message = "Your pet has passed away. Care cannot bring it back."
		return
	}
	if g.kind == "sleep" {
		a.putTamagotchiToSleep(g.index)
		after, _ := a.tamagotchiSnapshot()
		g.message = fmt.Sprintf("Well rested! Energy %d → %d\nHealth %d → %d | Happiness %d → %d", before.Energy, after.Energy, before.Health, after.Health, before.Happiness, after.Happiness)
	} else {
		a.stateMu.Lock()
		t := a.currentTamagotchi
		if t == nil || !t.IsAlive {
			a.stateMu.Unlock()
			g.message = "Your pet cannot be cared for right now."
			return
		}
		if (g.kind == "brush" || g.kind == "diaper") && g.kind != ageCareKind(*t) {
			a.stateMu.Unlock()
			g.message = "Your pet has grown: use " + ageCareLabel(*t) + " now."
			return
		}
		if g.kind == "diaper" {
			t.Diaper = 100
			t.LastChange = time.Now()
			g.message = fmt.Sprintf("Fresh diaper! Diaper %d → 100", before.Diaper)
		} else if g.kind == "wash" {
			t.Cleanliness = 100
			t.LastWash = time.Now()
			g.message = fmt.Sprintf("Squeaky clean! Cleanliness %d → 100", before.Cleanliness)
		} else {
			t.Grooming = 100
			t.LastBrush = time.Now()
			t.GroomTicks = 0
			g.message = fmt.Sprintf("Freshly brushed! Brushing %d → 100", before.Grooming)
		}
		a.resetRecoveryIfNeededLocked()
		// Hygiene stops neglect damage; washing alone does not heal existing illness.
		a.stateMu.Unlock()
		a.updateConfigFromState()
		a.addGameEvent("HYGIENE", g.message)
	}
}
