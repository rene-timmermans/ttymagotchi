package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func activateStatusRow(t *testing.T, a *App, prefix string) {
	t.Helper()
	a.TApp.SetRoot(a.TLayout, true)
	list := a.viewsList["status"]
	a.TApp.SetFocus(list)
	for i := 0; i < list.GetItemCount(); i++ {
		text, _ := list.GetItemText(i)
		if strings.HasPrefix(text, prefix) {
			list.SetCurrentItem(i)
			list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) { a.TApp.SetFocus(p) })
			return
		}
	}
	t.Fatalf("status row not found: %q", prefix)
}

func TestStatusEnterOpensRelatedMenusInEveryLanguage(t *testing.T) {
	for _, code := range languageCodes {
		for _, tc := range []struct{ format, section string }{
			{"Hunger: %s %s", feedSection}, {"Happiness: %s %s", playSection},
			{"Energy: %s %s", sleepSection}, {"Hygiene: ", hygieneSection},
			{"Health: %s %s", helpSection}, {"Recovery: fully healthy", helpSection},
		} {
			a := testApp(t)
			a.Config.App.Language = code
			a.refreshLanguage()
			before := *a.currentTamagotchi
			prefix := strings.Split(a.tr(tc.format), "%")[0]
			// Refreshing the live values must preserve both the selection and callback.
			a.generateStatusList(a.viewsList["status"])
			activateStatusRow(t, a, prefix)
			section, _ := a.TLayout.GetItem(0).(*tview.Pages).GetFrontPage()
			if section != tc.section || a.TApp.GetFocus() != a.viewsList[strings.ToLower(tc.section)] {
				t.Fatalf("%s %s did not focus %s", code, tc.format, tc.section)
			}
			if *a.currentTamagotchi != before || len(a.gameEvents) != 0 {
				t.Fatal("opening choices changed pet stats")
			}
			if tc.section == helpSection && a.viewsList["help"].GetCurrentItem() != 1 {
				t.Fatal("recovery explanation not selected")
			}
		}
	}
}

func TestStatusEnterStartsOnlyTheSelectedCare(t *testing.T) {
	for _, code := range languageCodes {
		for _, kind := range []string{"wash", "diaper", "brush"} {
			a := testApp(t)
			a.Config.App.Language = code
			if kind == "brush" {
				a.currentTamagotchi.Stage = "adult"
			}
			a.refreshLanguage()
			before := *a.currentTamagotchi
			prefix := a.tr(ageCareLabel(before)) + ": "
			if kind == "wash" {
				prefix = strings.Split(a.tr("Cleanliness: %s"), "%")[0]
			}
			activateStatusRow(t, a, prefix)
			if a.activity == nil || a.activity.kind != kind {
				t.Fatalf("%s %s did not start the expected animation", code, kind)
			}
			if *a.currentTamagotchi != before || len(a.gameEvents) != 0 {
				t.Fatal("care applied before animation completed")
			}
			a.returnToStatus()
			if len(a.gameEvents) != 1 {
				t.Fatal("care did not commit exactly once")
			}
			section, _ := a.TLayout.GetItem(0).(*tview.Pages).GetFrontPage()
			if section != statusSection {
				t.Fatal("care did not return to Status")
			}
		}
	}
}

func TestStatusAgeCareUsesCurrentStageAndRespectsDeath(t *testing.T) {
	a := testApp(t)
	// Simulate growth between drawing the Diaper row and pressing Enter.
	prefix := fmt.Sprintf("%s: ", a.tr(ageCareLabel(*a.currentTamagotchi)))
	a.currentTamagotchi.Stage = "toddler"
	activateStatusRow(t, a, prefix)
	if a.activity == nil || a.activity.kind != "brush" {
		t.Fatal("stale row used the wrong age care")
	}
	a.returnToStatus()
	a.currentTamagotchi.IsAlive = false
	a.currentTamagotchi.Health = 0
	a.generateStatusList(a.viewsList["status"])
	before := *a.currentTamagotchi
	activateStatusRow(t, a, strings.Split(a.tr("Cleanliness: %s"), "%")[0])
	if a.activity != nil || before != *a.currentTamagotchi {
		t.Fatal("status shortcut cared for a dead pet")
	}
}
