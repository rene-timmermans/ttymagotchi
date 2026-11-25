package app

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func confirmFullMeal(t *testing.T, a *App) {
	t.Helper()
	if a.activity == nil || a.activity.kind != "refuse_food" {
		t.Fatal("full pet did not refuse food")
	}
	a.tickActivity(a.activity.started.Add(4 * time.Second))
	key(a, tcell.KeyRight, 0)
	key(a, tcell.KeyEnter, 0)
	if a.activity == nil || a.activity.kind != "food" {
		t.Fatal("explicit consent did not start feeding")
	}
}

func TestFullPetRefusesAndDefaultChoiceStopsInEveryLanguage(t *testing.T) {
	for _, code := range languageCodes {
		a := testApp(t)
		a.Config.App.Language = code
		a.currentTamagotchi.Hunger = 0
		before := *a.currentTamagotchi
		a.startFeeding(0)
		g := a.activity
		if g == nil || g.kind != "refuse_food" {
			t.Fatal("missing refusal")
		}
		first := g.view.GetText(false)
		a.renderActivity(g.started.Add(250 * time.Millisecond))
		if first == g.view.GetText(false) || !strings.Contains(first, a.tr("No!")) {
			t.Fatal("No! is not localized/animated")
		}
		expected := fmt.Sprintf(a.tr("I'm full! Feeding anyway costs %d health."), 10)
		if !strings.Contains(first, expected) {
			t.Fatal("health cost not shown")
		}
		a.startFeeding(1)
		if a.activity.index != 0 {
			t.Fatal("repeated selection replaced pending food")
		}
		a.tickActivity(g.started.Add(time.Hour))
		a.finishFeeding() // Cannot bypass explicit consent through animation completion.
		if *a.currentTamagotchi != before || len(a.gameEvents) != 0 || g.phase != "choice" || g.input != 0 {
			t.Fatal("refusal auto-fed or did not default to stopping")
		}
		key(a, tcell.KeyEnter, 0)
		assertStatus(t, a)
		if *a.currentTamagotchi != before || len(a.gameEvents) != 0 {
			t.Fatal("stopping changed the pet")
		}
	}
}

func TestEscapeAndRepeatedEnterCannotForceFeed(t *testing.T) {
	for _, keys := range [][]tcell.Key{{tcell.KeyEscape}, {tcell.KeyCtrlS}, {tcell.KeyEnter, tcell.KeyEnter}} {
		a := testApp(t)
		a.currentTamagotchi.Hunger = 0
		before := *a.currentTamagotchi
		a.startFeeding(0)
		for _, code := range keys {
			key(a, code, 0)
		}
		assertStatus(t, a)
		if before != *a.currentTamagotchi || len(a.gameEvents) != 0 {
			t.Fatal("cancel/Enter bypassed consent")
		}
	}
}

func TestExplicitContinueAppliesEachFoodOnce(t *testing.T) {
	for i, food := range availableFoods {
		a := testApp(t)
		p := a.currentTamagotchi
		p.Hunger = 0
		before := *p
		a.startFeeding(i)
		confirmFullMeal(t, a)
		if *p != before || len(a.gameEvents) != 0 {
			t.Fatal("food applied before eating animation")
		}
		a.tickActivity(a.activity.started.Add(4 * time.Second))
		if p.Health != 100-overfeedingDamage(food, 0) || p.Weight != before.Weight+food.WeightGain || len(a.gameEvents) != 1 {
			t.Fatalf("wrong overfeeding effects for %s", food.Name)
		}
		after := *p
		a.tickActivity(time.Now().Add(time.Hour))
		if *p != after || len(a.gameEvents) != 1 {
			t.Fatal("meal applied twice")
		}
	}
}

func TestFatalWarningAndPendingPetDeath(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	p.Hunger = 0
	p.Health = 5
	a.startFeeding(0)
	if !strings.Contains(a.activity.view.GetText(false), a.tr("This would be fatal.")) {
		t.Fatal("fatal feeding is not disclosed")
	}
	confirmFullMeal(t, a)
	a.finishFeeding()
	if p.Health != 0 || p.IsAlive {
		t.Fatal("confirmed fatal feeding did not kill")
	}
	b := testApp(t)
	b.currentTamagotchi.Hunger = 0
	b.startFeeding(0)
	b.tickActivity(b.activity.started.Add(4 * time.Second))
	b.currentTamagotchi.Health = 0
	b.currentTamagotchi.IsAlive = false
	before := *b.currentTamagotchi
	key(b, tcell.KeyRune, '2')
	if b.activity != nil || *b.currentTamagotchi != before || len(b.gameEvents) != 0 {
		t.Fatal("pending meal fed or revived a dead pet")
	}
}

func TestHungerAndDamageAreRecheckedWhileChoosing(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	p.Hunger = 0
	a.startFeeding(0)
	a.tickActivity(a.activity.started.Add(4 * time.Second))
	p.Hunger = 20
	a.renderActivity(time.Now())
	if !strings.Contains(a.activity.view.GetText(false), a.tr("Your pet is hungry again. Feeding is safe now.")) {
		t.Fatal("warning did not follow current hunger")
	}
	key(a, tcell.KeyRune, '2')
	a.finishFeeding()
	if p.Health != 100 || p.Hunger != 0 {
		t.Fatal("stale warning caused incorrect damage")
	}
}

func TestFoodRefusalChoicesFitSmallTerminal(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)
	for _, code := range languageCodes {
		a := testApp(t)
		a.Config.App.Language = code
		p := a.currentTamagotchi
		p.Hunger = 0
		p.Health = 5
		p.Stage = "adult"
		p.Species = "dog"
		a.startFeeding(0)
		a.tickActivity(a.activity.started.Add(4 * time.Second))
		a.activity.view.SetRect(0, 0, 80, 24)
		a.activity.view.Draw(screen)
		screen.Show()
		cells, _, _ := screen.GetContents()
		var text strings.Builder
		for _, cell := range cells {
			text.WriteString(string(cell.Runes))
		}
		for _, key := range []string{"Stop feeding", "Feed anyway", "←/→: choose • Enter: confirm • Esc: stop", "This would be fatal."} {
			if !strings.Contains(text.String(), a.tr(key)) {
				t.Fatalf("%s choice text clipped: %s", code, key)
			}
		}
	}
}
