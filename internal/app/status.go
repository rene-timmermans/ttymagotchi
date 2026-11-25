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
		listStatus.AddItem(a.tr("No tamagotchi data available."), "", 0, nil)
		if a.spriteView != nil {
			a.spriteView.SetText("\n  ??\n")
		}
		return
	}

	// Status header
	status := a.tr("🟢 Alive")
	if t.Health <= 40 {
		status = a.tr("🤒 Sick — needs care")
	}
	if !t.IsAlive {
		status = a.tr("🔴 Dead")
	}
	listStatus.AddItem(fmt.Sprintf(a.tr("Status: %s"), status), "", 0, nil)

	// Basic info
	listStatus.AddItem(fmt.Sprintf(a.tr("Name: %s"), t.Name), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf(a.tr("Age: %d days"), t.Age), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf(a.tr("Stage: %s"), a.stageText(t)), "", 0, nil)
	listStatus.AddItem(a.tr("Species: ")+a.tr(speciesLabel(t)), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf(a.tr("Weight: %.1f grams"), t.Weight), "", 0, nil)

	// Stats with visual bars
	listStatus.AddItem("", "", 0, nil) // Empty line
	listStatus.AddItem(a.tr("=== STATS (↑/↓ + Enter) ==="), "", 0, nil)

	// Hunger bar
	hungerBar := a.createProgressBar(t.Hunger, 100)
	hungerColor := "🍽️"
	if t.Hunger > 80 {
		hungerColor = "😫"
	} else if t.Hunger > 60 {
		hungerColor = "🤤"
	}
	listStatus.AddItem(fmt.Sprintf(a.tr("Hunger: %s %s"), hungerColor, hungerBar), "", 0, func() { a.openStatusSection(feedSection) })

	// Happiness bar
	happinessBar := a.createProgressBar(t.Happiness, 100)
	happinessColor := "😊"
	if t.Happiness < 20 {
		happinessColor = "😢"
	} else if t.Happiness < 50 {
		happinessColor = "😐"
	}
	listStatus.AddItem(fmt.Sprintf(a.tr("Happiness: %s %s"), happinessColor, happinessBar), "", 0, func() { a.openStatusSection(playSection) })

	// Health bar
	healthBar := a.createProgressBar(t.Health, 100)
	healthColor := healthIcon(t.Health)
	listStatus.AddItem(fmt.Sprintf(a.tr("Health: %s %s"), healthColor, healthBar), "", 0, a.openRecoveryHelp)

	// Energy bar
	energyBar := a.createProgressBar(t.Energy, 100)
	energyColor := "⚡"
	if t.Energy < 20 {
		energyColor = "😴"
	} else if t.Energy < 50 {
		energyColor = "😪"
	}
	listStatus.AddItem(fmt.Sprintf(a.tr("Energy: %s %s"), energyColor, energyBar), "", 0, func() { a.openStatusSection(sleepSection) })

	listStatus.AddItem(a.tr("Hygiene: ")+"🧴 "+a.createProgressBar(hygieneLevel(t), 100), "", 0, func() { a.openStatusSection(hygieneSection) })
	listStatus.AddItem(fmt.Sprintf(a.tr("Cleanliness: %s"), "🧼 "+a.createProgressBar(t.Cleanliness, 100)), "", 0, func() { a.startCare("wash", 0) })
	listStatus.AddItem(fmt.Sprintf("%s: %s", a.tr(ageCareLabel(t)), ageCareIcon(t)+" "+a.createProgressBar(ageCareLevel(t), 100)), "", 0, a.startStatusAgeCare)

	listStatus.AddItem(a.recoveryStatus(t), "", 0, a.openRecoveryHelp)

	// Last actions
	listStatus.AddItem("", "", 0, nil) // Empty line
	listStatus.AddItem(a.tr("=== LAST ACTIONS ==="), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf(a.tr("Last Fed: %s"), t.LastFed.Format("15:04")), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf(a.tr("Last Play: %s"), t.LastPlay.Format("15:04")), "", 0, nil)
	listStatus.AddItem(fmt.Sprintf(a.tr("Last Sleep: %s"), t.LastSleep.Format("15:04")), "", 0, nil)

	// Created date
	listStatus.AddItem("", "", 0, nil) // Empty line
	listStatus.AddItem(a.tr("=== INFO ==="), "", 0, nil)
	if t.Created.IsZero() {
		listStatus.AddItem(a.tr("Created: Unknown"), "", 0, nil)
		listStatus.AddItem(a.tr("Time Alive: Unknown"), "", 0, nil)
	} else {
		listStatus.AddItem(fmt.Sprintf(a.tr("Created: %s"), t.Created.Format("2006-01-02 15:04")), "", 0, nil)
		listStatus.AddItem(fmt.Sprintf(a.tr("Time Alive: %s"), time.Since(t.Created).Round(time.Second)), "", 0, nil)
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
		view.SetBorder(true).SetTitle(a.tr("Current Form"))
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

func healthIcon(health int) string {
	if health < 30 {
		return "💔"
	}
	if health < 70 {
		return "💛"
	}
	return "💚"
}

// Status uses the same menus as the shortcuts when a choice is required.
func (a *App) openStatusSection(section string) {
	a.goToSection(section, nil)
	if list := a.viewsList[strings.ToLower(section)]; list != nil {
		a.TApp.SetFocus(list)
	}
}

func (a *App) openRecoveryHelp() {
	a.openStatusSection(helpSection)
	a.viewsList["help"].SetCurrentItem(1) // CARE & HEALTH section.
}

func (a *App) startStatusAgeCare() {
	// Resolve the action at activation time, including growth since the last draw.
	if pet, ok := a.tamagotchiSnapshot(); ok {
		a.startCare(ageCareKind(pet), 0)
	}
}
