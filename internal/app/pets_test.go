package app

import (
	"strings"
	"testing"
	"time"

	"github.com/ezeoleaf/termagotchi/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

func TestFiveSpeciesTenDistinctFormsAndHiddenEgg(t *testing.T) {
	eggs := map[string]bool{}
	for _, species := range petSpecies {
		forms := map[string]bool{}
		p := *newDefaultTamagotchi("Test")
		p.Species = species
		for _, stage := range growthStages {
			p.Stage = stage
			sprite := renderTamagotchiSprite(p)
			if forms[sprite] {
				t.Fatalf("duplicate %s form at %s", species, stage)
			}
			forms[sprite] = true
			if stage == "egg" {
				eggs[sprite] = true
				if speciesLabel(p) != "Mystery egg" {
					t.Fatal("species leaked")
				}
			}
			for _, line := range strings.Split(sprite, "\n") {
				if uniseg.StringWidth(line) > 36 {
					t.Fatalf("%s %s too wide for status panel", species, stage)
				}
			}
			for _, kind := range []string{"food", "sleep", "wash", "brush"} {

				first := strings.Join(actionPose(p, kind, 0), "\n")
				next := strings.Join(actionPose(p, kind, 1), "\n")
				if first == next {
					t.Fatalf("%s %s %s body does not animate", species, stage, kind)
				}
				for frame := 0; frame < 16; frame++ {
					scene := renderActionSprite(p, kind, 0, time.Duration(frame)*250*time.Millisecond)
					if len(strings.Split(scene, "\n")) != sceneHeight {
						t.Fatal("unstable height")
					}
					for _, row := range strings.Split(scene, "\n") {
						if uniseg.StringWidth(row) != sceneWidth {
							t.Fatalf("%s %s %s frame %d width", species, stage, kind, frame)
						}
					}
				}
			}
		}
		if len(forms) != 10 {
			t.Fatal("not ten growth forms")
		}
	}
	if len(eggs) != 1 {
		t.Fatal("eggs reveal the species")
	}
}

func TestSpeciesPersistAndRandomNewPets(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		p := newDefaultTamagotchi("Test")
		if !knownSpecies(p.Species) || p.Stage != "egg" {
			t.Fatal("invalid new pet")
		}
		seen[p.Species] = true
	}
	if len(seen) != 5 {
		t.Fatal("random selection did not include all species")
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	a := testApp(t)
	a.Config = cfg
	a.currentTamagotchi.Species = "penguin"
	a.updateConfigFromState()
	cfg.App.LastLogin = time.Now()
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	saved, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	saved.App.CurrentLogin = saved.App.LastLogin
	b := testApp(t)
	b.Config = saved
	b.initializeStateFromConfig()
	if b.currentTamagotchi.Species != "penguin" {
		t.Fatal("species rerolled on reload")
	}
	b.Config.Tamagotchi.Species = ""
	b.initializeStateFromConfig()
	if !knownSpecies(b.currentTamagotchi.Species) {
		t.Fatal("legacy pet not migrated")
	}
}

func TestGrowthAndBrushingAtEveryAge(t *testing.T) {
	for i, stage := range growthStages {
		a := testApp(t)
		p := a.currentTamagotchi
		p.Age = growthDays[i]
		a.updateStageLocked(p.Stage)
		if p.Stage != stage {
			t.Fatalf("wrong stage on day %d", p.Age)
		}
		p.Grooming = 0
		if needsMet(*p) || hygieneLevel(*p) != 0 {
			t.Fatal("neglected brushing ignored")
		}
		a.startCare("brush", 0)
		if a.activity == nil {
			t.Fatalf("brushing unavailable at %s", stage)
		}
		a.finishCare()
		if p.Grooming != 100 || p.GroomTicks != 0 {
			t.Fatal("brushing did not restore coat")
		}
	}
}

func TestBrushingDecaysSlowlyAndPersists(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	p.Hunger = 0
	a.advanceTime(20 * a.updateInterval)
	if p.Grooming != 100 || p.GroomTicks != 20 {
		t.Fatal("brushing declined too early")
	}
	a.advanceTime(a.updateInterval)
	if p.Grooming != 99 || p.GroomTicks != 0 {
		t.Fatal("brushing did not decline after 21 periods")
	}
	if 21*100*defaultUpdateInterval != 7*24*time.Hour {
		t.Fatal("brushing duration is not one normal week")
	}
	a.updateConfigFromState()
	if a.Config.Tamagotchi.Grooming != 99 || a.Config.Tamagotchi.GroomTicks != 0 {
		t.Fatal("brushing not saved")
	}
}

func TestLargePetsAndControlsFitStandardTerminal(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)
	screenText := func() string {
		screen.Show()
		cells, _, _ := screen.GetContents()
		var b strings.Builder
		for _, c := range cells {
			b.WriteString(string(c.Runes))
		}
		return b.String()
	}
	for _, species := range petSpecies {
		a := testApp(t)
		a.currentTamagotchi.Species = species
		a.currentTamagotchi.Stage = "adult"
		for _, kind := range []string{"food", "sleep", "wash", "brush"} {
			if kind == "food" {
				a.startFeeding(0)
			} else {
				a.startCare(kind, 0)
			}
			a.activity.view.SetRect(0, 0, 80, 24)
			a.activity.view.Draw(screen)
			if text := screenText(); !strings.Contains(text, "Esc: Status") {
				t.Fatalf("%s %s controls clipped", species, kind)
			}
			a.closeActivity()
		}
		for _, game := range []int{0, 4, 5} {
			a.playWithTamagotchi(game)
			a.activity.view.SetRect(0, 0, 80, 24)
			a.activity.view.Draw(screen)
			if text := screenText(); !strings.Contains(text, "Press Enter to start") {
				t.Fatalf("%s game %d start hint clipped", species, game)
			}
			a.closeActivity()
		}
	}
}
