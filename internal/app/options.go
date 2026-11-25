package app

import (
	"fmt"
	"math"
	"time"

	"github.com/ezeoleaf/termagotchi/internal/config"
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
	a.updateConfigFromState()
	a.Config.App.LastLogin = time.Now()
	a.Config.App.CurrentLogin = a.Config.App.LastLogin
	if err := config.SaveConfig(a.Config); err != nil {
		a.addGameEvent("ERROR", fmt.Sprintf("Could not save speed: %v", err))
	}
	a.requestRefresh()
}

func (a *App) generateOptionsList(list *tview.List) {
	defer preserveSelection(list)()
	list.Clear()
	speed := a.Config.App.SpeedPercent
	list.AddItem(fmt.Sprintf("Speed: %d%%   ← slower / faster →   (+ / -)", speed), "", 0, nil)
	list.AddItem(fmt.Sprintf("Hunger 0 → 100: %s", formatDuration(hungerDuration(speed))), "", 0, nil)
	list.AddItem("Changes also affect energy, hygiene and neglect damage.", "", 0, nil)
	list.AddItem("Offline time counts. Animations keep their normal speed.", "", 0, nil)
	list.AddItem("", "", 0, nil)
	for _, value := range []int{0, 25, 50, 75, 100} {
		v := value
		label := fmt.Sprintf("%d%% — %s from full to starving", v, formatDuration(hungerDuration(v)))
		if v == 50 {
			label += " (normal: about 3 full meals/day)"
		}
		list.AddItem(label, "", 0, func() { a.setSpeed(v); a.generateOptionsList(list) })
	}
	list.AddItem("", "", 0, nil)
	list.AddItem("Left/right or +/-: 1% | Page Up/Down: 10%", "", 0, nil)
	list.AddItem("Settings save immediately; 50% is the default.", "", 0, nil)
}

func (a *App) optionsPage() (string, tview.Primitive) {
	list := getList()
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
