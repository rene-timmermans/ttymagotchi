package app

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Activity state belongs exclusively to the UI event loop.
type activity struct {
	view                           *tview.TextView
	focus                          tview.Primitive
	kind, phase                    string
	index, round, score, input     int
	sequence                       []int
	started, deadline              time.Time
	message                        string
	target, lane, step, posX, posY int
	word                           string
	choices                        []string
}

var arrows = []string{"↑", "→", "↓", "←"}

func gameKind(index int) string {
	return []string{"timing", "runner", "notes", "drawing", "memory", "arrows", "reading", "seek"}[index]
}

func (a *App) openActivity(kind string, index int) {
	v := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	v.SetBorder(true).SetTitle(a.tr(" ttymagotchi • Activity ")).SetBorderColor(tcell.ColorLightCyan)
	a.uiMu.Lock()
	a.activityActive = true
	a.uiMu.Unlock()
	a.activity = &activity{view: v, focus: a.TApp.GetFocus(), kind: kind, index: index, phase: "ready"}
	a.TApp.SetRoot(v, true).SetFocus(v)
}

func (a *App) closeActivity() {
	focus := a.activity.focus
	a.activity = nil
	a.uiMu.Lock()
	a.activityActive = false
	a.uiMu.Unlock()
	a.TApp.SetRoot(a.TLayout, true)
	if focus != nil {
		a.TApp.SetFocus(focus)
	}
	a.requestRefresh()
}

func (a *App) startGame(index int) {
	a.openActivity(gameKind(index), index)
	a.renderActivity(time.Now())
}

func (a *App) startFeeding(index int) {
	if index < 0 || index >= len(availableFoods) || a.activity != nil || a.modal != nil {
		return
	}
	t, ok := a.tamagotchiSnapshot()
	if !ok || !t.IsAlive {
		return
	}
	kind := "food"
	if t.Hunger == 0 {
		kind = "refuse_food"
	}
	a.openActivity(kind, index)
	a.activity.started = time.Now()
	a.activity.phase = "playing"
	a.renderActivity(time.Now())
}

func (a *App) beginRound(now time.Time) {
	g := a.activity
	g.phase, g.input, g.started = "playing", 0, now
	if a.beginExtraGame(now) {
		return
	}
	g.deadline = now.Add(5 * time.Second)
	if g.kind != "timing" {
		n := 5
		if g.kind == "memory" {
			n = 3 + g.round
			g.phase = "show"
		}
		g.sequence = make([]int, n)
		for i := range g.sequence {
			g.sequence[i] = rand.IntN(4)
		}
		g.deadline = now.Add(time.Duration(n+3) * time.Second)
	}
}

func (a *App) finishRound(success bool, now time.Time) {
	g := a.activity
	if success {
		g.score++
		g.message = a.tr("[green]Great job![-]")
	} else {
		g.message = a.tr("[yellow]Try again next round![-]")
	}
	g.round++
	if g.round == 3 {
		g.phase = "result"
		g.message = a.rewardGame(g.index, g.score, 3)
	} else {
		g.phase = "between"
		g.deadline = now.Add(900 * time.Millisecond)
	}
}

// The marker bounces across 25 cells; the five central cells form the target.
func timingPosition(elapsed time.Duration) int {
	step := int(elapsed/(65*time.Millisecond)) % 48
	if step > 24 {
		step = 48 - step
	}
	return step
}

func (a *App) activityInput(event *tcell.EventKey) *tcell.EventKey {
	g := a.activity
	// Let tview handle Ctrl+C, including its normal save-on-exit path.
	if event.Key() == tcell.KeyCtrlC {
		return event
	}
	if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlS {
		a.returnToStatus()
		return nil
	}
	now := time.Now()
	a.tickActivity(now)
	if g.kind == "refuse_food" {
		return a.foodRefusalInput(event, now)
	}
	if g.phase == "result" {
		if event.Key() == tcell.KeyEnter || event.Rune() == ' ' {
			a.closeActivity()
		}
		return nil
	}
	if g.phase == "ready" {
		if event.Key() == tcell.KeyEnter {
			a.beginRound(now)
		}
	} else if g.phase == "playing" {
		if a.extraGameInput(event, now) {
		} else if g.kind == "timing" && event.Rune() == ' ' {
			pos := timingPosition(now.Sub(g.started))
			a.finishRound(pos >= 10 && pos <= 14, now)
		} else if g.kind == "memory" || g.kind == "arrows" {
			direction := -1
			switch event.Key() {
			case tcell.KeyUp:
				direction = 0
			case tcell.KeyRight:
				direction = 1
			case tcell.KeyDown:
				direction = 2
			case tcell.KeyLeft:
				direction = 3
			}
			if direction < 0 {
				switch event.Rune() {
				case 'w', 'W':
					direction = 0
				case 'd', 'D':
					direction = 1
				case 's', 'S':
					direction = 2
				case 'a', 'A':
					direction = 3
				}
			}
			if direction >= 0 {
				if direction != g.sequence[g.input] {
					a.finishRound(false, now)
				} else {
					g.input++
					if g.input == len(g.sequence) {
						a.finishRound(true, now)
					}
				}
			}
		}
	}
	a.renderActivity(now)
	return nil
}

func (a *App) finishFeeding() {
	g := a.activity
	if g.kind != "food" || g.phase == "result" {
		return
	}
	before, _ := a.tamagotchiSnapshot()
	a.feedTamagotchi(g.index)
	after, _ := a.tamagotchiSnapshot()
	g.phase = "result"
	if !before.IsAlive {
		g.message = a.tr("Your pet cannot eat right now.")
		return
	}
	g.message = fmt.Sprintf(a.tr("Hunger %d → %d | Happiness %d → %d\nEnergy %d → %d | Health %d → %d"), before.Hunger, after.Hunger, before.Happiness, after.Happiness, before.Energy, after.Energy, before.Health, after.Health)
	if after.Health < before.Health {
		g.message = a.tr("Overfed! Your pet feels sick.\n") + g.message
	} else {
		g.message = a.tr("Yum! ") + g.message
	}
	if !after.IsAlive {
		g.message += a.tr("\nYour tamagotchi has passed away...")
	}
}

func (a *App) tickActivity(now time.Time) {
	g := a.activity
	if g == nil {
		return
	}
	if g.kind == "refuse_food" {
		if g.phase == "playing" && now.Sub(g.started) >= 3*time.Second {
			g.phase = "choice"
		}
	} else if g.kind == "refuse" {
		if now.Sub(g.started) >= 3*time.Second {
			g.phase = "result"
		}
	} else if isCare(g.kind) {
		if g.phase != "result" && now.Sub(g.started) >= 4*time.Second {
			a.finishCare()
		}
	} else if g.kind == "food" {
		if g.phase != "result" && now.Sub(g.started) >= 3*time.Second {
			a.finishFeeding()
		}
	} else if a.tickExtraGame(now) {
	} else {
		switch g.phase {
		case "show":
			if now.Sub(g.started) >= time.Duration(len(g.sequence))*650*time.Millisecond+500*time.Millisecond {
				g.phase = "playing"
				g.deadline = now.Add(time.Duration(len(g.sequence)+3) * time.Second)
			}
		case "playing":
			if !now.Before(g.deadline) {
				a.finishRound(false, now)
			}
		case "between":
			if !now.Before(g.deadline) {
				a.beginRound(now)
			}
		}
	}
	a.renderActivity(now)
}

func (a *App) renderActivity(now time.Time) {
	g := a.activity
	t, _ := a.tamagotchiSnapshot()
	title, body, hint := "", "", a.tr("Esc: Status (cancel game) • Ctrl+C: quit")
	sprite := renderTamagotchiSprite(t)
	if g.kind == "refuse_food" {
		title, body, hint, sprite = a.foodRefusalView(t, now)
	} else if g.kind == "refuse" {
		title = "[red]" + a.tr("No!") + "[-]"
		body = a.tr("Not sleepy! Energy must be 60% or lower.")
		sprite = renderActionSprite(t, "refuse", g.index, now.Sub(g.started))
		hint = a.tr("Esc: Status • Ctrl+C: quit")
	} else if isCare(g.kind) {
		hint = a.tr("Esc: Status (finish care) • Ctrl+C: quit")
		sprite = renderActionSprite(t, g.kind, g.index, now.Sub(g.started))
		switch g.kind {
		case "sleep":
			title = a.tr(sleepOptions[g.index].Name)
			body = a.tr("Sleeping peacefully...")
		case "wash":
			title = a.tr("A bubbly bath")
			body = a.tr("Washing away the dirt...")
		case "diaper":
			title = a.tr("Diaper")
			body = a.tr("Changing into a fresh diaper...")
		case "brush":
			title = a.tr("Brushing")
			body = a.tr("Brushing fur and feathers...")
		}
	} else if g.kind == "food" {
		title = a.tr(availableFoods[g.index].Name)
		hint = a.tr("Esc: Status (finish care) • Ctrl+C: quit")
		sprite = renderActionSprite(t, g.kind, g.index, now.Sub(g.started))
		if now.Sub(g.started) < 1500*time.Millisecond {
			body = a.tr("Dinner is coming!")
		} else {
			body = a.tr("Munch!")
		}

	} else {
		title = a.tr(availableGames[g.index].Name)
		switch g.kind {
		case "runner", "notes", "drawing", "reading", "seek":
			body = a.extraGameText(now)
		case "timing":
			body = a.tr("Press SPACE when the marker is in the green zone.")
			if g.phase == "playing" {
				var bar strings.Builder
				pos := timingPosition(now.Sub(g.started))
				for i := 0; i < 25; i++ {
					color := "gray"
					if i >= 10 && i <= 14 {
						color = "green"
					}
					cell := "─"
					if i == pos {
						cell = "●"
					}
					fmt.Fprintf(&bar, "[%s]%s[-]", color, cell)
				}
				body += "\n\n" + bar.String()
			}
		case "memory":
			body = a.tr("Remember the arrows, then repeat them.\nUse arrow keys or WASD.")
			if g.phase == "show" {
				body = a.tr("Watch carefully!\n\n")
				active := int(now.Sub(g.started) / (650 * time.Millisecond))
				for i, d := range g.sequence {
					if i == active {
						body += "[yellow]" + arrows[d] + "[-]  "
					} else {
						body += "·  "
					}
				}
			} else if g.phase == "playing" {
				body = fmt.Sprintf(a.tr("Repeat the hidden sequence!\n\n%d / %d arrows"), g.input, len(g.sequence))
			}
		case "arrows":
			body = a.tr("Follow the arrows in order.\nUse arrow keys or WASD.")
			if g.phase == "playing" {
				body += "\n\n"
				for i, d := range g.sequence {
					color := "white"
					if i < g.input {
						color = "green"
					} else if i == g.input {
						color = "yellow"
					}
					body += "[" + color + "]" + arrows[d] + "[-]  "
				}
			}
		}
		if g.phase == "ready" {
			body += a.tr("\n\nThree rounds • Score determines rewards\nEnergy is spent only on completion.\n\n[green]Press Enter to start[-]")
		}
		if g.phase == "playing" {
			body += fmt.Sprintf(a.tr("\n\nTime left: %.1fs"), maxWeight(0, g.deadline.Sub(now).Seconds()))
		}
		if g.phase == "between" {
			body = g.message + a.tr("\n\nNext round...")
		}
		title += fmt.Sprintf(a.tr(" • Round %d/3 • Score %d"), min(3, g.round+1), g.score)
	}
	if g.phase == "result" {
		sprite = renderTamagotchiSprite(t)
		body = a.tr("[green]All done![-]\n\n") + tview.Escape(g.message)
		if g.kind == "refuse" {
			body = a.tr("Not sleepy! Energy must be 60% or lower.")
		}
		hint = a.tr("Enter: menu • Esc: Status • Ctrl+C: quit")
	}
	if g.kind != "food" && !isCare(g.kind) {
		body = strings.ReplaceAll(body, "\n\n", "\n")
	}
	g.view.SetText("[::b]" + title + "[::-]\n" + sprite + "\n" + body + "\n[gray]" + hint + "[-]")
}
