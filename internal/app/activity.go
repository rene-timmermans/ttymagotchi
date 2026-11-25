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
	view                       *tview.TextView
	focus                      tview.Primitive
	kind, phase                string
	index, round, score, input int
	sequence                   []int
	started, deadline          time.Time
	message                    string
}

var arrows = []string{"↑", "→", "↓", "←"}

func gameKind(index int) string {
	switch index {
	case 4, 6, 7:
		return "memory"
	case 3, 5:
		return "arrows"
	default:
		return "timing"
	}
}

func (a *App) openActivity(kind string, index int) {
	v := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	v.SetBorder(true).SetTitle(" Termagotchi • Activity ").SetBorderColor(tcell.ColorLightCyan)
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
	a.openActivity("food", index)
	a.activity.started = time.Now()
	a.activity.phase = "playing"
	a.renderActivity(time.Now())
}

func (a *App) beginRound(now time.Time) {
	g := a.activity
	g.phase, g.input, g.started = "playing", 0, now
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
		g.message = "[green]Great job![-]"
	} else {
		g.message = "[yellow]Try again next round![-]"
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
	now := time.Now()
	a.tickActivity(now)
	if event.Key() == tcell.KeyEscape {
		if g.kind == "food" && g.phase != "result" {
			a.finishFeeding()
		} else {
			a.closeActivity()
		}
		return nil
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
		if g.kind == "timing" && event.Rune() == ' ' {
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
	if g.phase == "result" {
		return
	}
	before, _ := a.tamagotchiSnapshot()
	a.feedTamagotchi(g.index)
	after, _ := a.tamagotchiSnapshot()
	g.phase = "result"
	if !after.IsAlive {
		g.message = "Your pet cannot eat right now."
		return
	}
	g.message = fmt.Sprintf("Yum! Hunger %d → %d | Happiness %d → %d\nEnergy %d → %d", before.Hunger, after.Hunger, before.Happiness, after.Happiness, before.Energy, after.Energy)
}

func (a *App) tickActivity(now time.Time) {
	g := a.activity
	if g == nil {
		return
	}
	if g.kind == "food" {
		if g.phase != "result" && now.Sub(g.started) >= 3*time.Second {
			a.finishFeeding()
		}
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
	title, body, hint := "", "", "Esc: cancel • Ctrl+C: quit"
	sprite := renderTamagotchiSprite(t)
	if g.kind == "food" {
		title = availableFoods[g.index].Name
		hint = "Esc: skip animation • Ctrl+C: quit"
		elapsed := now.Sub(g.started)
		if elapsed < 1500*time.Millisecond {
			gap := max(0, 12-int(elapsed/(125*time.Millisecond)))
			icon := strings.Fields(availableFoods[g.index].Name)[0]
			body = icon + strings.Repeat(" ", gap) + " →\nDinner is coming!"
		} else {
			if int(elapsed/(250*time.Millisecond))%2 == 0 {
				body = "( ^ o ^ )\nMunch!"
			} else {
				body = "( ^ - ^ )\nNom nom..."
			}
		}
	} else {
		title = availableGames[g.index].Name
		switch g.kind {
		case "timing":
			body = "Press SPACE when the marker is in the green zone."
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
			body = "Remember the arrows, then repeat them.\nUse arrow keys or WASD."
			if g.phase == "show" {
				body = "Watch carefully!\n\n"
				active := int(now.Sub(g.started) / (650 * time.Millisecond))
				for i, d := range g.sequence {
					if i == active {
						body += "[yellow]" + arrows[d] + "[-]  "
					} else {
						body += "·  "
					}
				}
			} else if g.phase == "playing" {
				body = fmt.Sprintf("Repeat the hidden sequence!\n\n%d / %d arrows", g.input, len(g.sequence))
			}
		case "arrows":
			body = "Follow the arrows in order.\nUse arrow keys or WASD."
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
			body += "\n\nThree rounds • Score determines rewards\nEnergy is spent only on completion.\n\n[green]Press Enter to start[-]"
		}
		if g.phase == "playing" {
			body += fmt.Sprintf("\n\nTime left: %.1fs", maxWeight(0, g.deadline.Sub(now).Seconds()))
		}
		if g.phase == "between" {
			body = g.message + "\n\nNext round..."
		}
		title += fmt.Sprintf(" • Round %d/3 • Score %d", min(3, g.round+1), g.score)
	}
	if g.phase == "result" {
		body = "[green]All done![-]\n\n" + tview.Escape(g.message)
		if g.kind == "food" {
			t.Happiness = 100
			sprite = renderTamagotchiSprite(t)
		}
		hint = "Enter / Esc: back • Ctrl+C: quit"
	}
	g.view.SetText("\n[::b]" + title + "[::-]\n" + sprite + "\n" + body + "\n\n[gray]" + hint + "[-]")
}
