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
		list.AddItem(a.tr("No tamagotchi available."), "", 0, nil)
		return
	}
	if !t.IsAlive {
		list.AddItem(a.tr("Your tamagotchi has passed away... 💔"), "", 0, nil)
		return
	}
	list.AddItem(a.tr("=== HYGIENE ==="), "", 0, nil)
	list.AddItem(a.tr("Wash — a bubbly bath"), "", 0, func() { a.startCare("wash", 0) })
	if needsDiaper(t) {
		list.AddItem(a.tr("Diaper — change into a fresh diaper"), "", 0, func() { a.startCare("diaper", 0) })
	} else {
		list.AddItem(a.tr("Brushing — care for fur and feathers"), "", 0, func() { a.startCare("brush", 0) })
	}
	list.AddItem("", "", 0, nil)
	list.AddItem(a.tr("Hygiene: ")+a.createProgressBar(hygieneLevel(t), 100), "", 0, nil)
	list.AddItem(a.tr("Cleanliness: ")+a.createProgressBar(t.Cleanliness, 100), "", 0, nil)
	list.AddItem(a.tr(ageCareLabel(t))+": "+a.createProgressBar(ageCareLevel(t), 100), "", 0, nil)
	list.AddItem(fmt.Sprintf(a.tr("Health: %d/100"), t.Health), "", 0, nil)
	list.AddItem("", "", 0, nil)
	list.AddItem(a.tr("At 20% or below, neglected hygiene damages health."), "", 0, nil)
	list.AddItem(a.tr("Closer to 0% = more damage. Each active need counts independently."), "", 0, nil)
	if needsDiaper(t) {
		list.AddItem(a.tr("Diapers are needed through stage 3; Brushing starts at stage 4."), "", 0, nil)
	} else {
		list.AddItem(a.tr("Brushing declines very slowly: 100 → 0 in ~7 days at 50% speed."), "", 0, nil)
	}
	list.AddItem(a.tr("Baths restore cleanliness; the second action restores age-specific care."), "", 0, nil)
	careName, careAt := a.tr("Last brushing"), t.LastBrush
	if needsDiaper(t) {
		careName, careAt = a.tr("Last diaper change"), t.LastChange
	}
	for _, action := range []struct {
		name string
		at   time.Time
	}{{a.tr("Last bath"), t.LastWash}, {careName, careAt}} {
		when := a.tr("Not yet")
		if !action.at.IsZero() {
			when = action.at.Format(a.tr("02 Jan 15:04"))
		}
		list.AddItem(a.tr(action.name)+": "+when, "", 0, nil)
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
	if kind == "sleep" && t.Energy > 60 {
		a.openActivity("refuse", index)
		a.activity.phase = "playing"
		a.activity.started = time.Now()
		a.activity.message = a.tr("Not sleepy! Energy must be 60% or lower.")
		a.renderActivity(time.Now())
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
		g.message = a.tr("Your pet has passed away. Care cannot bring it back.")
		return
	}
	if g.kind == "sleep" {
		a.putTamagotchiToSleep(g.index)
		after, _ := a.tamagotchiSnapshot()
		g.message = fmt.Sprintf(a.tr("Well rested! Energy %d → %d\nHealth %d → %d | Happiness %d → %d"), before.Energy, after.Energy, before.Health, after.Health, before.Happiness, after.Happiness)
	} else {
		a.stateMu.Lock()
		t := a.currentTamagotchi
		if t == nil || !t.IsAlive {
			a.stateMu.Unlock()
			g.message = a.tr("Your pet cannot be cared for right now.")
			return
		}
		if (g.kind == "brush" || g.kind == "diaper") && g.kind != ageCareKind(*t) {
			a.stateMu.Unlock()
			g.message = fmt.Sprintf(a.tr("Your pet has grown: use %s now."), a.tr(ageCareLabel(*t)))
			return
		}
		if g.kind == "diaper" {
			t.Diaper = 100
			t.LastChange = time.Now()
			a.addLocalizedEvent("HYGIENE", "Fresh diaper! Diaper %d → 100", before.Diaper)
			g.message = fmt.Sprintf(a.tr("Fresh diaper! Diaper %d → 100"), before.Diaper)
		} else if g.kind == "wash" {
			t.Cleanliness = 100
			t.LastWash = time.Now()
			a.addLocalizedEvent("HYGIENE", "Squeaky clean! Cleanliness %d → 100", before.Cleanliness)
			g.message = fmt.Sprintf(a.tr("Squeaky clean! Cleanliness %d → 100"), before.Cleanliness)
		} else {
			t.Grooming = 100
			t.LastBrush = time.Now()
			t.GroomTicks = 0
			a.addLocalizedEvent("HYGIENE", "Freshly brushed! Brushing %d → 100", before.Grooming)
			g.message = fmt.Sprintf(a.tr("Freshly brushed! Brushing %d → 100"), before.Grooming)
		}
		a.resetRecoveryIfNeededLocked()
		// Hygiene stops neglect damage; washing alone does not heal existing illness.
		a.stateMu.Unlock()
		a.updateConfigFromState()
	}
}
