package app

import (
	"fmt"
	"math"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Piecewise geometric interpolation keeps both halves of the scale useful.
func hungerDuration(speed int) time.Duration {
	speed = max(0, min(100, speed))
	if speed == 0 {
		return 7 * 24 * time.Hour
	}
	if speed == 50 {
		return 8 * time.Hour
	}
	if speed == 100 {
		return 5 * time.Minute
	}
	low, high, fraction := float64(7*24*time.Hour), float64(8*time.Hour), float64(speed)/50
	if speed > 50 {
		low, high, fraction = float64(8*time.Hour), float64(5*time.Minute), float64(speed-50)/50
	}
	return time.Duration(low * math.Pow(high/low, fraction))
}

func preserveSelection(list *tview.List) func() {
	index := list.GetCurrentItem()
	return func() {
		if n := list.GetItemCount(); n > 0 {
			list.SetCurrentItem(min(index, n-1))
		}
	}
}

func (a *App) setSpeed(speed int) {
	speed = max(0, min(100, speed))
	a.stateMu.Lock()
	interval := hungerDuration(speed) / 100
	// Preserve the fraction of a hunger point already elapsed.
	if a.updateInterval > 0 {
		a.timeAccumulator = time.Duration(float64(a.timeAccumulator) / float64(a.updateInterval) * float64(interval))
	}
	a.updateInterval = interval
	a.stateMu.Unlock()
	a.Config.App.SpeedPercent = speed
	a.saveSettings()
	a.requestRefresh()
}

func (a *App) generateOptionsList(list *tview.List) {
	defer preserveSelection(list)()
	list.Clear()
	list.AddItem(fmt.Sprintf(a.tr("Language: %s"), languageNames[normalizeLanguage(a.Config.App.Language)]), "", 0, func() { a.cycleLanguage(1) })
	for _, language := range languageCodes {
		code := language
		list.AddItem("  "+languageNames[code], "", 0, func() { a.setLanguage(code) })
	}
	list.AddItem("", "", 0, nil)
	speed := a.Config.App.SpeedPercent
	list.AddItem(fmt.Sprintf(a.tr("Speed: %d%%   ← slower / faster →   (+ / -)"), speed), "", 0, nil)
	list.AddItem(fmt.Sprintf(a.tr("Hunger 0 → 100: %s"), a.duration(hungerDuration(speed))), "", 0, nil)
	list.AddItem(a.tr("Changes also affect energy, hygiene and neglect damage."), "", 0, nil)
	list.AddItem(a.tr("Offline time counts. Animations keep their normal speed."), "", 0, nil)
	list.AddItem("", "", 0, nil)
	for _, value := range []int{0, 25, 50, 75, 100} {
		v := value
		label := fmt.Sprintf(a.tr("%d%% — %s from full to starving"), v, a.duration(hungerDuration(v)))
		if v == 50 {
			label += a.tr(" (normal: about 3 full meals/day)")
		}
		list.AddItem(label, "", 0, func() { a.setSpeed(v); a.generateOptionsList(list) })
	}
	list.AddItem("", "", 0, nil)
	list.AddItem(a.tr("Left/right or +/-: 1% | Page Up/Down: 10%"), "", 0, nil)
	list.AddItem(a.tr("Settings save immediately; 50% is the default."), "", 0, nil)
}

func (a *App) optionsPage() (string, tview.Primitive) {
	list := getList()
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if list.GetCurrentItem() <= 4 && (event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyRight) {
			delta := 1
			if event.Key() == tcell.KeyLeft {
				delta = -1
			}
			a.cycleLanguage(delta)
			return nil
		}
		delta := 0
		switch event.Key() {
		case tcell.KeyLeft:
			delta = -1
		case tcell.KeyRight:
			delta = 1
		case tcell.KeyPgDn:
			delta = -10
		case tcell.KeyPgUp:
			delta = 10
		}
		switch event.Rune() {
		case '-':
			delta = -1
		case '+', '=':
			delta = 1
		}
		if delta != 0 {
			a.setSpeed(a.Config.App.SpeedPercent + delta)
			a.generateOptionsList(list)
			return nil
		}
		return event
	})
	a.viewsList["options"] = list
	a.generateOptionsList(list)
	return optionsSection, list
}
