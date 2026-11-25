package app

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rene-timmermans/ttymagotchi/internal/config"
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"
)

func TestAllCatalogsCompleteAndFormatsMatch(t *testing.T) {
	verbs := regexp.MustCompile(`%[-+0-9.]*[sdvf%]`)
	en := translations["en"]
	required := append([]string{footerText, "Current Form"}, helpText...)
	for _, food := range availableFoods {
		required = append(required, food.Name)
	}
	for _, game := range availableGames {
		required = append(required, game.Name)
	}
	for _, sleep := range sleepOptions {
		required = append(required, sleep.Name)
	}
	required = append(required, growthStages...)
	files, _ := filepath.Glob("*.go")
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			index := -1
			if sel.Sel.Name == "tr" {
				index = 0
			}
			if sel.Sel.Name == "addLocalizedEvent" {
				index = 1
			}
			if index >= 0 && len(call.Args) > index {
				if literal, ok := call.Args[index].(*ast.BasicLit); ok && literal.Kind == token.STRING {
					key, _ := strconv.Unquote(literal.Value)
					required = append(required, key)
				}
			}
			return true
		})
	}
	for _, code := range languageCodes {
		catalog := translations[code]
		if len(catalog) != len(en) {
			t.Fatalf("%s has missing keys", code)
		}
		for _, key := range required {
			if _, ok := catalog[key]; !ok {
				t.Errorf("%s missing %q", code, key)
			}
		}
		for key, value := range catalog {
			if !reflect.DeepEqual(verbs.FindAllString(key, -1), verbs.FindAllString(value, -1)) {
				t.Errorf("%s incompatible placeholders: %q → %q", code, key, value)
			}
			if value == "" {
				t.Errorf("%s empty translation: %q", code, key)
			}
		}
		for _, line := range strings.Split(catalog[footerText], "\n") {
			if uniseg.StringWidth(line) > 80 {
				t.Errorf("%s footer too wide", code)
			}
		}
	}
}

func TestLanguageAppliesImmediatelyAndPersists(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	a := testApp(t)
	a.Config = cfg
	a.currentTamagotchi.Name = "Energy"
	a.addLocalizedEvent("RESTART", "Started a new tamagotchi named %s! 🥚", "Energy")
	a.addLocalizedEvent("FEED", "Fed %s! Hunger -%d, Happiness +%d, Health -%d", translatedArgument(availableFoods[0].Name), 20, 5, 0)
	for _, code := range languageCodes {
		a.setLanguage(code)
		title, _ := a.viewsList["status"].GetItemText(1)
		if title != fmt.Sprintf(a.tr("Name: %s"), "Energy") {
			t.Fatalf("%s translated a pet name", code)
		}
		if !strings.Contains(a.eventMessage(a.gameEvents[0]), "Energy") {
			t.Fatal("pet name changed in log")
		}
		if !strings.Contains(a.eventMessage(a.gameEvents[1]), a.tr(availableFoods[0].Name)) {
			t.Fatal("old event did not change language")
		}
		if !strings.Contains(a.infoView.GetText(true), strings.Split(a.tr(footerText), "\n")[1]) {
			t.Fatal("footer did not update")
		}
		saved, err := config.LoadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if saved.App.Language != code {
			t.Fatal("language not persisted")
		}
	}
	// Unknown saved codes use English; legacy files default to English too.
	if err := os.WriteFile(filepath.Join(cfg.App.SaveDirectory, "config.yml"), []byte("app:\n  language: invalid\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err = config.LoadConfig()
	if err != nil || cfg.App.Language != "en" {
		t.Fatal("invalid language has no fallback")
	}
}

func TestRefusalAndCareLocalizedWithoutChangingStats(t *testing.T) {
	for _, code := range languageCodes {
		a := testApp(t)
		a.Config.App.Language = code
		a.currentTamagotchi.Energy = 61
		before := *a.currentTamagotchi
		a.startCare("sleep", 0)
		if a.activity.kind != "refuse" {
			t.Fatal("sleep not refused")
		}
		a.renderActivity(a.activity.started)
		first := a.activity.view.GetText(false)
		a.renderActivity(a.activity.started.Add(250 * time.Millisecond))
		if first == a.activity.view.GetText(false) {
			t.Fatal("refusal body does not animate")
		}
		if !strings.Contains(first, a.tr("No!")) || !strings.Contains(first, a.tr("Not sleepy! Energy must be 60% or lower.")) {
			t.Fatal("refusal not localized")
		}
		a.tickActivity(a.activity.started.Add(4 * time.Second))
		if before != *a.currentTamagotchi || len(a.gameEvents) != 0 {
			t.Fatal("refused sleep changed stats")
		}
		a.returnToStatus()
		a.currentTamagotchi.Energy = 60
		a.startCare("sleep", 0)
		if a.activity.kind != "sleep" {
			t.Fatal("sleep rejected at exactly 60")
		}
		a.finishCare()
		if a.currentTamagotchi.Energy != 80 {
			t.Fatal("permitted sleep not applied")
		}
	}
}

func TestLocalizedGameControlsFitStandardTerminal(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)
	for _, code := range languageCodes {
		for i := range availableGames {
			a := testApp(t)
			a.Config.App.Language = code
			a.currentTamagotchi.Stage = "adult"
			a.currentTamagotchi.Species = "dog"
			a.playWithTamagotchi(i)
			for phase := 0; phase < 2; phase++ {
				if phase == 1 {
					key(a, tcell.KeyEnter, 0)
				}
				a.activity.view.SetRect(0, 0, 80, 24)
				a.activity.view.Draw(screen)
				screen.Show()
				cells, _, _ := screen.GetContents()
				var text strings.Builder
				for _, c := range cells {
					text.WriteString(string(c.Runes))
				}
				if !strings.Contains(text.String(), a.tr("Esc: Status (cancel game) • Ctrl+C: quit")) {
					t.Errorf("%s game %d phase %d controls clipped", code, i, phase)
				}
			}
		}
	}
}

func TestReadingUsesTheSelectedLanguage(t *testing.T) {
	for _, code := range languageCodes {
		a := testApp(t)
		a.Config.App.Language = code
		a.playWithTamagotchi(6)
		key(a, tcell.KeyEnter, 0)
		g := a.activity
		correct := string(unicode.ToUpper([]rune(g.word)[0]))
		if g.choices[g.target] != correct {
			t.Fatal("reading answer does not match translated word")
		}
		key(a, tcell.KeyRune, rune('1'+g.target))
		if g.score != 1 || g.phase != "between" {
			t.Fatalf("%s reading failed", code)
		}
	}
}

func TestDutchOptionSelectionAndLanguageArrows(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	a := testApp(t)
	a.Config = cfg
	list := a.viewsList["options"]
	list.SetCurrentItem(len(languageCodes)) // Last language row: Nederlands.
	list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) { a.TApp.SetFocus(p) })
	if a.Config.App.Language != "nl" {
		t.Fatal("Dutch option did not select Dutch")
	}
	text, _ := list.GetItemText(0)
	if text != "Taal: Nederlands" {
		t.Fatalf("Dutch heading: %q", text)
	}
	for _, step := range []struct {
		key  tcell.Key
		want string
	}{{tcell.KeyRight, "en"}, {tcell.KeyLeft, "nl"}} {
		list.InputHandler()(tcell.NewEventKey(step.key, 0, tcell.ModNone), func(p tview.Primitive) { a.TApp.SetFocus(p) })
		if a.Config.App.Language != step.want || a.Config.App.SpeedPercent != 50 {
			t.Fatal("language arrows changed speed or failed to wrap")
		}
	}
	saved, err := config.LoadConfig()
	if err != nil || saved.App.Language != "nl" {
		t.Fatal("Dutch preference did not survive reload")
	}
}
