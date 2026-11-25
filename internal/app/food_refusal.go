package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

// A refused meal stays pending until the player explicitly chooses to continue.
// Enter defaults to stopping; repeated Enter presses cannot force-feed the pet.
func (a *App) foodRefusalInput(event *tcell.EventKey, now time.Time) *tcell.EventKey {
	g := a.activity
	if pet, ok := a.tamagotchiSnapshot(); !ok || !pet.IsAlive {
		a.returnToStatus()
		return nil
	}
	if g.phase == "playing" {
		if event.Key() == tcell.KeyEnter {
			g.phase = "choice"
		}
	} else if g.phase == "choice" {
		confirm := false
		switch event.Key() {
		case tcell.KeyLeft, tcell.KeyUp:
			g.input = 0
		case tcell.KeyRight, tcell.KeyDown:
			g.input = 1
		case tcell.KeyEnter:
			confirm = true
		}
		switch event.Rune() {
		case '1':
			g.input = 0
			confirm = true
		case '2':
			g.input = 1
			confirm = true
		}
		if confirm {
			if g.input == 0 {
				a.returnToStatus()
				return nil
			}
			// Reuse the activity to preserve its original menu focus on return.
			g.kind = "food"
			g.phase = "playing"
			g.started = now
			g.message = ""
			g.input = 0
		}
	}
	a.renderActivity(now)
	return nil
}

func (a *App) foodRefusalView(pet Tamagotchi, now time.Time) (title, body, hint, sprite string) {
	g := a.activity
	title = "[red]" + a.tr("No!") + "[-]"
	damage := overfeedingDamage(availableFoods[g.index], pet.Hunger)
	body = fmt.Sprintf(a.tr("I'm full! Feeding anyway costs %d health."), damage)
	if damage == 0 {
		title = a.tr("Feed?")
		body = a.tr("Your pet is hungry again. Feeding is safe now.")
	}
	if damage > 0 && damage >= pet.Health {
		body += "\n" + a.tr("This would be fatal.")
	}
	sprite = renderActionSprite(pet, "refuse", g.index, now.Sub(g.started))
	if damage == 0 {
		sprite = renderTamagotchiSprite(pet)
	}
	// Remove empty scene rows, keeping the entire larger pet and both choices
	// visible even in a terminal with 24 rows.
	rows := strings.Split(sprite, "\n")
	for len(rows) > 0 && strings.TrimSpace(rows[len(rows)-1]) == "" {
		rows = rows[:len(rows)-1]
	}
	sprite = strings.Join(rows, "\n")
	hint = a.tr("Enter: show choices • Esc: stop feeding")
	if g.phase == "choice" {
		choices := []string{"1: " + a.tr("Stop feeding"), "2: " + a.tr("Feed anyway")}
		choices[g.input] = "[yellow]› " + choices[g.input] + "[-]"
		body += "\n" + strings.Join(choices, "    ")
		hint = a.tr("←/→: choose • Enter: confirm • Esc: stop")
	}
	return
}
