package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// RunArtPreview shows the live game renderers without loading a configuration,
// starting the simulation or saving a pet. All state belongs to this preview.
func RunArtPreview() error {
	terminal := tview.NewApplication()
	view := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetWrap(false)
	header := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetWrap(false)
	footer := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetWrap(false)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(view, 0, 1, true).
		AddItem(footer, 3, 0, false)
	layout.SetBorder(true).SetTitle(" ttymagotchi • Voorbeeldmodus ")
	species, stage, mood, action, food := 0, len(growthStages)-1, 1, 0, 0
	speciesNames := []string{"Kat", "Hond", "Konijn", "Cavia", "Pinguïn / Tux"}
	moodNames := []string{"Vrolijk", "Tevreden", "Neutraal", "Bezorgd", "Verdrietig", "Huilend", "Gestorven"}
	moodValues := []int{100, 70, 50, 30, 10, 0, 0}
	started := time.Now()
	paused := false
	tired := false
	compact := false
	elapsed := time.Duration(0)
	render := func() {
		if !paused {
			elapsed = time.Since(started)
		}
		pet := Tamagotchi{Species: petSpecies[species], Stage: growthStages[stage], Happiness: moodValues[mood], Health: 100, Energy: 100, IsAlive: mood < 6}
		if tired {
			pet.Energy = 10
		}
		if !pet.IsAlive {
			pet.Health = 0
		}
		pet.compactArt = compact
		kinds := []string{"idle", "food", "wash", ageCareKind(pet), "sleep", "refuse"}
		labels := []string{"Rust", "Eten", "Wassen", "Borstelen", "Slapen", "Nee!"}
		if needsDiaper(pet) {
			labels[3] = "Luier verschonen"
		}
		art := renderActionSprite(pet, kinds[action], food, elapsed%(4*time.Second))
		if action == 0 {
			art = strings.Join(actionPose(pet, "idle", int(elapsed/(400*time.Millisecond))), "\n")
		}
		rows := strings.Split(art, "\n")
		for len(rows) > 0 && strings.TrimSpace(rows[0]) == "" {
			rows = rows[1:]
		}
		for len(rows) > 0 && strings.TrimSpace(rows[len(rows)-1]) == "" {
			rows = rows[:len(rows)-1]
		}
		state := ""
		if paused {
			state = " • gepauzeerd"
		}
		if tired {
			state += " • moe"
		}
		if pet.compactArt {
			state += " • compact"
		}
		if !pet.compactArt {
			state += " • detail"
		}
		heading := fmt.Sprintf("%s • %s (%d/10) • %s • %s%s", speciesNames[species], pet.Stage, stage+1, moodNames[mood], labels[action], state)
		controls := "1–5: dier • ↑/↓: fase • ←/→: stemming • Tab/Enter: actie\nC: compact/detail • PgUp/PgDn: scroll • Esc/Q: sluiten\nF: eten • T: moe/fit • Spatie: pauze • eigen save blijft intact."
		header.SetText(heading)
		view.SetText(strings.Join(rows, "\n"))
		footer.SetText(controls)
	}
	terminal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		reset := true
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			terminal.Stop()
			return nil
		case tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyHome, tcell.KeyEnd:
			return event
		case tcell.KeyUp:
			stage = (stage + 1) % len(growthStages)
		case tcell.KeyDown:
			stage = (stage + len(growthStages) - 1) % len(growthStages)
		case tcell.KeyLeft:
			mood = (mood + 1) % len(moodNames)
		case tcell.KeyRight:
			mood = (mood + len(moodNames) - 1) % len(moodNames)
		case tcell.KeyTab, tcell.KeyEnter:
			action = (action + 1) % 6
		case tcell.KeyF24:
			reset = false // Animation timer, processed on the UI thread.
		case tcell.KeyRune:
			switch r := event.Rune(); {
			case r >= '1' && r <= '5':
				species = int(r - '1')
			case r == 'q' || r == 'Q':
				terminal.Stop()
				return nil
			case r == 'c' || r == 'C':
				compact = !compact
				view.ScrollToBeginning()
			case r == 't' || r == 'T':
				tired = !tired
			case r == 'f' || r == 'F':
				action = 1
				food = (food + 1) % len(availableFoods)
			case r == ' ':
				if paused {
					started = time.Now().Add(-elapsed)
				} else {
					elapsed = time.Since(started)
				}
				paused = !paused
				reset = false
			default:
				reset = false
			}
		default:
			reset = false
		}
		if reset {
			view.ScrollToBeginning()
			started = time.Now()
			elapsed = 0
		}
		render()
		return nil
	})
	view.SetRect(0, 0, 80, 24)
	render()
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(125 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				terminal.QueueEvent(tcell.NewEventKey(tcell.KeyF24, 0, tcell.ModNone))
			}
		}
	}()
	return terminal.SetRoot(layout, true).SetFocus(view).EnableMouse(true).Run()
}
