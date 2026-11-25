package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rene-timmermans/ttymagotchi/internal/config"
	"github.com/rivo/tview"
)

func testApp(t *testing.T) *App {
	t.Helper()
	a := &App{TApp: tview.NewApplication(), Config: &config.Config{}, viewsList: make(map[string]*tview.List), currentTamagotchi: newDefaultTamagotchi("Test"), updateInterval: defaultUpdateInterval}
	pages, info := a.getPagesInfo()
	a.TLayout = tview.NewFlex().AddItem(pages, 0, 1, true).AddItem(info, 1, 0, false)
	return a
}

func key(a *App, k tcell.Key, r rune) { a.activityInput(tcell.NewEventKey(k, r, tcell.ModNone)) }

func TestAllGamesRequireCompletion(t *testing.T) {
	for i, game := range availableGames {
		t.Run(game.Name, func(t *testing.T) {
			a := testApp(t)
			before, _ := a.tamagotchiSnapshot()
			a.playWithTamagotchi(i)
			if a.activity == nil {
				t.Fatal("no minigame")
			}
			after, _ := a.tamagotchiSnapshot()
			if before != after {
				t.Fatal("selection changed stats")
			}
			key(a, tcell.KeyEscape, 0)
			after, _ = a.tamagotchiSnapshot()
			if before != after || len(a.gameEvents) != 0 {
				t.Fatal("cancel changed stats")
			}
			a.playWithTamagotchi(i)
			key(a, tcell.KeyEnter, 0)
			for round := 0; round < 3; round++ {
				g := a.activity
				if g.phase == "between" {
					a.tickActivity(g.deadline.Add(time.Millisecond))
				}
				winRound(t, a)

			}
			g := a.activity
			if g.phase != "result" || g.score != 3 {
				t.Fatalf("result: %+v", g)
			}
			after, _ = a.tamagotchiSnapshot()
			if after.Happiness != min(100, before.Happiness+game.Happiness) || after.Energy != before.Energy+game.Energy {
				t.Fatal("wrong reward")
			}
			a.tickActivity(time.Now().Add(time.Hour))
			if len(a.gameEvents) != 1 {
				t.Fatal("duplicate reward")
			}
			key(a, tcell.KeyEnter, 0)
			if a.activity != nil {
				t.Fatal("did not return")
			}
		})
	}
}

func TestFailureTimeoutAndGuards(t *testing.T) {
	a := testApp(t)
	a.playWithTamagotchi(4)
	key(a, tcell.KeyEnter, 0)
	a.tickActivity(a.activity.started.Add(5 * time.Second))
	wrong := (a.activity.sequence[0] + 1) % 4
	key(a, tcell.KeyRune, []rune{'w', 'd', 's', 'a'}[wrong])
	if a.activity.round != 1 || a.activity.score != 0 {
		t.Fatal("wrong input accepted")
	}
	a.tickActivity(a.activity.deadline)
	a.tickActivity(a.activity.started.Add(5 * time.Second))
	a.tickActivity(a.activity.deadline)
	a.tickActivity(a.activity.deadline)
	a.tickActivity(a.activity.started.Add(5 * time.Second))
	a.tickActivity(a.activity.deadline)
	if a.activity.phase != "result" || a.activity.score != 0 || a.currentTamagotchi.Happiness != 50 {
		t.Fatal("timeout rewarded")
	}
	key(a, tcell.KeyEscape, 0)
	a.currentTamagotchi.Energy = 9
	a.playWithTamagotchi(0)
	if a.activity != nil {
		t.Fatal("tired pet allowed")
	}
	a.currentTamagotchi.Energy = 100
	a.currentTamagotchi.IsAlive = false
	a.playWithTamagotchi(0)
	a.startFeeding(0)
	if a.activity != nil {
		t.Fatal("dead pet allowed")
	}
}

func TestFeedingAnimationCommitsOnce(t *testing.T) {
	for _, skip := range []bool{false, true} {
		a := testApp(t)
		before, _ := a.tamagotchiSnapshot()
		a.startFeeding(0)
		g := a.activity
		if a.currentTamagotchi.Hunger != before.Hunger {
			t.Fatal("fed before animation")
		}
		a.startFeeding(1) // Repeated activation must not replace the meal.
		a.tickActivity(a.activity.started.Add(time.Second))
		if skip {
			key(a, tcell.KeyEscape, 0)
		} else {
			a.tickActivity(a.activity.started.Add(4 * time.Second))
		}
		a.tickActivity(time.Now().Add(time.Hour))
		if a.currentTamagotchi.Hunger != 30 || len(a.gameEvents) != 1 {
			t.Fatal("missing or duplicate meal")
		}
		if g.phase != "result" || !strings.Contains(g.message, "50 → 30") {
			t.Fatal("missing result")
		}
		if a.activity != nil {
			key(a, tcell.KeyEnter, 0)
		}
		if a.activity != nil {
			t.Fatal("not back at menu")
		}
	}
}

func TestActivityDrawsInSmallTerminal(t *testing.T) {
	a := testApp(t)
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)
	a.startFeeding(0)
	a.activity.view.SetRect(0, 0, 80, 23)
	a.activity.view.Draw(screen)
	screen.Show()
	cells, _, _ := screen.GetContents()
	var visible strings.Builder
	for _, c := range cells {
		visible.WriteString(string(c.Runes))
	}
	if !strings.Contains(visible.String(), "Dinner is coming!") {
		t.Fatal("animation not visible")
	}
	a.closeActivity()
	a.playWithTamagotchi(0)
	key(a, tcell.KeyEnter, 0)
	a.activity.view.SetRect(0, 0, 80, 23)
	a.activity.view.Draw(screen)
	screen.Show()
	cells, _, _ = screen.GetContents()
	visible.Reset()
	for _, c := range cells {
		visible.WriteString(string(c.Runes))
	}
	if !strings.Contains(visible.String(), "Time left:") {
		t.Fatal("game not visible")
	}
}

func winRound(t *testing.T, a *App) {
	t.Helper()
	g := a.activity
	switch g.kind {
	case "timing":
		g.started = time.Now().Add(-12 * 65 * time.Millisecond)
		key(a, tcell.KeyRune, ' ')
	case "memory", "arrows":
		if g.phase == "show" {
			a.tickActivity(g.started.Add(5 * time.Second))
		}
		for _, d := range append([]int(nil), g.sequence...) {
			key(a, []tcell.Key{tcell.KeyUp, tcell.KeyRight, tcell.KeyDown, tcell.KeyLeft}[d], 0)
		}
	case "runner":
		for g.phase == "playing" {
			step := g.step
			safe := (g.sequence[step] + 1) % 3
			g.started = time.Now().Add(-time.Duration(step) * 700 * time.Millisecond)
			for g.lane < safe {
				key(a, tcell.KeyRight, 0)
			}
			for g.lane > safe {
				key(a, tcell.KeyLeft, 0)
			}
			g.started = time.Now().Add(-time.Duration(step+1)*700*time.Millisecond - 10*time.Millisecond)
			a.tickActivity(time.Now())
		}
	case "notes":
		for _, note := range append([]int(nil), g.sequence...) {
			key(a, tcell.KeyRune, rune('1'+note))
		}
	case "drawing":
		for _, move := range []struct {
			key tcell.Key
			n   int
		}{{tcell.KeyRight, 4}, {tcell.KeyDown, 2}, {tcell.KeyLeft, 4}, {tcell.KeyUp, 2}} {
			for n := 0; n < move.n; n++ {
				key(a, move.key, 0)
			}
		}
	case "reading":
		key(a, tcell.KeyRune, rune('1'+g.target))
	case "seek":
		a.tickActivity(g.started.Add(1300 * time.Millisecond))
		a.tickActivity(g.started.Add(time.Duration(len(g.sequence)) * 800 * time.Millisecond))
		key(a, tcell.KeyRune, rune('1'+g.target))
	default:
		t.Fatalf("no driver for %s", g.kind)
	}
}
