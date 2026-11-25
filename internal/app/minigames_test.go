package app

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestEightDistinctGames(t *testing.T) {
	kinds := map[string]bool{}
	for i := range availableGames {
		kind := gameKind(i)
		if kinds[kind] {
			t.Fatalf("duplicate mechanic: %s", kind)
		}
		kinds[kind] = true
	}
}
func TestExtraGameFailuresAndTimeouts(t *testing.T) {
	for _, i := range []int{1, 2, 3, 6, 7} {
		a := testApp(t)
		a.playWithTamagotchi(i)
		key(a, tcell.KeyEnter, 0)
		g := a.activity
		switch g.kind {
		case "runner":
			g.lane = g.sequence[0]
			a.tickActivity(g.started.Add(710 * time.Millisecond))
		case "notes":
			key(a, tcell.KeyRune, rune('1'+(g.sequence[0]+1)%3))
		case "drawing":
			key(a, tcell.KeyRight, 0)
			key(a, tcell.KeyDown, 0)
		case "reading":
			key(a, tcell.KeyRune, rune('1'+(g.target+1)%3))
		case "seek":
			a.tickActivity(g.started.Add(1300 * time.Millisecond))
			a.tickActivity(g.started.Add(time.Duration(len(g.sequence)) * 800 * time.Millisecond))
			key(a, tcell.KeyRune, rune('1'+(g.target+1)%3))
		}
		if g.phase != "between" || g.score != 0 || g.round != 1 {
			t.Fatalf("failure not handled for %s", g.kind)
		}
		a.tickActivity(g.deadline.Add(time.Millisecond))
		if g.kind == "seek" {
			a.tickActivity(g.started.Add(1300 * time.Millisecond))
			a.tickActivity(g.started.Add(time.Duration(len(g.sequence)) * 800 * time.Millisecond))
		}
		if g.kind == "runner" {
			g.lane = g.sequence[0]
		}
		a.tickActivity(g.deadline.Add(time.Second))
		if g.round != 2 {
			t.Fatalf("timeout not handled for %s", g.kind)
		}
	}
}
func TestCannotBypassSleepRefusal(t *testing.T) {
	a := testApp(t)
	a.currentTamagotchi.Energy = 61
	before := *a.currentTamagotchi
	a.putTamagotchiToSleep(0)
	if *a.currentTamagotchi != before {
		t.Fatal("direct sleep bypassed energy limit")
	}
}
