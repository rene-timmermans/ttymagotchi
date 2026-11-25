package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
)

func (a *App) generateStatusList(listStatus *tview.List) {
	defer preserveSelection(listStatus)()
	listStatus.Clear()

	t, ok := a.tamagotchiSnapshot()
	if !ok {
		listStatus.AddItem("No tamagotchi data available.", "", 0, nil)
		if a.spriteView != nil {
			a.spriteView.SetText("\n  ??\n")
		}
		return
	}

	// Status header
	status := "🟢 Alive"
	if t.Health <= 40 {
		status = "🤒 Sick — needs care"
	}
	if !t.IsAlive {
		status = "🔴 Dead"
	}
	listStatus.AddItem(fmt.Sprintf("Status: %s", status), "", 0, nil)

	// Basic info
	listStatus.AddItem(fmt.Sprintf("Name: %s", t.Name), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf("Age: %d days", t.Age), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf("Stage: %s", stageLabel(t)), "", 0, nil)
	listStatus.AddItem("Species: "+speciesLabel(t), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf("Weight: %.1f grams", t.Weight), "", 0, nil)

	// Stats with visual bars
	listStatus.AddItem("", "", 0, nil) // Empty line
	listStatus.AddItem("=== STATS ===", "", 0, nil)

	// Hunger bar
	hungerBar := a.createProgressBar(t.Hunger, 100)
	hungerColor := "🟢"
	if t.Hunger > 80 {
		hungerColor = "🔴"
	} else if t.Hunger > 60 {
		hungerColor = "🟡"
	}
	listStatus.AddItem(fmt.Sprintf("Hunger: %s %s", hungerColor, hungerBar), "", 0, nil)

	// Happiness bar
	happinessBar := a.createProgressBar(t.Happiness, 100)
	happinessColor := "😊"
	if t.Happiness < 20 {
		happinessColor = "😢"
	} else if t.Happiness < 50 {
		happinessColor = "😐"
	}
	listStatus.AddItem(fmt.Sprintf("Happiness: %s %s", happinessColor, happinessBar), "", 0, nil)

	// Health bar
	healthBar := a.createProgressBar(t.Health, 100)
	healthColor := "🟢"
	if t.Health < 30 {
		healthColor = "🔴"
	} else if t.Health < 70 {
		healthColor = "🟡"
	}
	listStatus.AddItem(fmt.Sprintf("Health: %s %s", healthColor, healthBar), "", 0, nil)

	// Energy bar
	energyBar := a.createProgressBar(t.Energy, 100)
	energyColor := "⚡"
	if t.Energy < 20 {
		energyColor = "😴"
	} else if t.Energy < 50 {
		energyColor = "😪"
	}
	listStatus.AddItem(fmt.Sprintf("Energy: %s %s", energyColor, energyBar), "", 0, nil)

	listStatus.AddItem("Hygiene: "+a.createProgressBar(hygieneLevel(t), 100), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf("Cleanliness: %s", a.createProgressBar(t.Cleanliness, 100)), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf("Brushing: %s", a.createProgressBar(t.Grooming, 100)), "", 0, nil)

	listStatus.AddItem(a.recoveryStatus(t), "", 0, nil)

	// Last actions
	listStatus.AddItem("", "", 0, nil) // Empty line
	listStatus.AddItem("=== LAST ACTIONS ===", "", 0, nil)
	listStatus.AddItem(fmt.Sprintf("Last Fed: %s", t.LastFed.Format("15:04")), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf("Last Play: %s", t.LastPlay.Format("15:04")), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf("Last Sleep: %s", t.LastSleep.Format("15:04")), "", 0, nil)

	// Created date
	listStatus.AddItem("", "", 0, nil) // Empty line
	listStatus.AddItem("=== INFO ===", "", 0, nil)
	if t.Created.IsZero() {
		listStatus.AddItem("Created: Unknown", "", 0, nil)
		listStatus.AddItem("Time Alive: Unknown", "", 0, nil)
	} else {
		listStatus.AddItem(fmt.Sprintf("Created: %s", t.Created.Format("2006-01-02 15:04")), "", 0, nil)
		listStatus.AddItem(fmt.Sprintf("Time Alive: %s", time.Since(t.Created).Round(time.Second)), "", 0, nil)
	}
}

func (a *App) updateSpriteView(view *tview.TextView) {
	t, ok := a.tamagotchiSnapshot()
	if !ok {
		view.SetText("\n  ??\n")
		return
	}

	if knownSpecies(t.Species) && growthIndex(t.Stage) > 0 {
		view.SetText(strings.Join(animalPose(t, "idle", int(time.Now().UnixMilli()/400)), "\n"))
	} else if growthIndex(t.Stage) == 0 && t.IsAlive {
		frame := int(time.Now().UnixMilli() / 400)
		pose := eggPose(t, "idle", frame)
		if frame%2 == 1 {
			for i := range pose {
				pose[i] = " " + pose[i]
			}
		}
		view.SetText(strings.Join(pose, "\n"))
	} else {
		view.SetText(renderTamagotchiSprite(t))
	}
}

func (a *App) createProgressBar(current, max int) string {
	const barLength = 20
	if max <= 0 {
		return "[0%] " + strings.Repeat("░", barLength)
	}
	if current < 0 {
		current = 0
	}
	if current > max {
		current = max
	}
	eighths := current * barLength * 8 / max
	full, fraction := eighths/8, eighths%8
	bar := strings.Repeat("█", full)
	if fraction > 0 {
		bar += []string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}[fraction]
	}
	empty := barLength - full
	if fraction > 0 {
		empty--
	}
	bar += strings.Repeat("░", empty)
	return fmt.Sprintf("[%d%%] %s", current*100/max, bar)
}

func (a *App) statusPage() (title string, content tview.Primitive) {
	listStatus := a.viewsList["status"]
	if listStatus == nil {
		listStatus = getList()
		a.viewsList["status"] = listStatus
	}

	a.generateStatusList(listStatus)

	if a.spriteView == nil {
		view := tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetDynamicColors(true)
		view.SetBorder(true).SetTitle("Current Form")
		a.spriteView = view
	}
	a.updateSpriteView(a.spriteView)

	layout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(listStatus, 0, 1, true).
		AddItem(a.spriteView, 0, 1, false)

	title = statusSection
	return title, layout
}

func renderTamagotchiSprite(t Tamagotchi) string {
	if growthIndex(t.Stage) == 0 {
		return strings.Join(eggPose(t, "idle", 0), "\n")
	}
	if knownSpecies(t.Species) {
		return strings.Join(animalPose(t, "idle", 0), "\n")
	}
	mood := spriteMood(t.Happiness, t.Health, t.Energy, t.IsAlive)
	stage := t.Stage
	if stage == "" {
		stage = "egg"
	}

	switch stage {
	case "egg":
		return eggSprites[mood]
	case "baby":
		return babySprites[mood]
	case "child":
		return childSprites[mood]
	case "teen":
		return teenSprites[mood]
	default:
		return adultSprites[mood]
	}
}

func spriteMood(happiness, health, energy int, alive bool) string {
	if !alive || health <= 0 {
		return "dead"
	}
	switch {
	case happiness > 75 && health > 75 && energy > 60:
		return "happy"
	case happiness < 25 || health <= 40 || energy < 20:
		return "sad"
	default:
		return "neutral"
	}
}

var eggSprites = map[string]string{
	"dead": `
  ⭕
 /XX\
 \__/
`,
	"sad": `
  ⭕
 /..\
 \__/
`,
	"happy": `
  ⭕
 /^^\
 \__/
`,
	"neutral": `
  ⭕
 /--\
 \__/
`,
}

var babySprites = map[string]string{
	"dead": `
   __
 _(xx)_
(      )
 \_/\_/
`,
	"sad": `
   __
 _(..)_ 
(  -- )
 \_/\_/
`,
	"happy": `
   __
 _(^^)_ 
(  \/ )
 \_/\_/
`,
	"neutral": `
   __
 _(--)_ 
(  \/ )
 \_/\_/
`,
}

var childSprites = map[string]string{
	"dead": `
  /\_/\
 ( x x )
 /  ^  \
 \__=__/
`,
	"sad": `
  /\_/\
 ( - - )
 /  ^  \
 \__~__/
`,
	"happy": `
  /\_/\
 ( ^ ^ )
 /  v  \
 \__~__/
`,
	"neutral": `
  /\_/\
 ( o o )
 /  v  \
 \__~__/
`,
}

var teenSprites = map[string]string{
	"dead": `
   /\_/\
  ( x x )
  /| ^ |\
 /_|___|_\
    |_|
`,
	"sad": `
   /\_/\
  ( - - )
  /| ^ |\
 /_|___|_\
    |_|
`,
	"happy": `
   /\_/\
  ( ^ ^ )
  /| v |\
 /_|___|_\
    |_|
`,
	"neutral": `
   /\_/\
  ( o o )
  /| v |\
 /_|___|_\
    |_|
`,
}

var adultSprites = map[string]string{
	"dead": `
   /\___/\
  ( x   x )
  /|  ^  |\
 /_| --- |_\
    /   \ 
   _\   /_
`,
	"sad": `
   /\___/\
  ( -   - )
  /|  ^  |\
 /_| ___ |_\
    /   \ 
   _\   /_
`,
	"happy": `
   /\___/\
  ( ^   ^ )
  /|  v  |\
 /_| ___ |_\
    /   \ 
   _\   /_
`,
	"neutral": `
   /\___/\
  ( o   o )
  /|  v  |\
 /_| ___ |_\
    /   \ 
   _\   /_
`,
}
