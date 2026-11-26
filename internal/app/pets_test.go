package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rene-timmermans/ttymagotchi/internal/config"
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
				if uniseg.StringWidth(line) != petArtSize(p).width {
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
					if len(strings.Split(scene, "\n")) != actionSceneSize(p).height {
						t.Fatal("unstable height")
					}
					for _, row := range strings.Split(scene, "\n") {
						if uniseg.StringWidth(row) != actionSceneSize(p).width {
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

func TestAgeAppropriateCareAtEveryStage(t *testing.T) {
	for i, stage := range growthStages {
		a := testApp(t)
		p := a.currentTamagotchi
		p.Age = growthDays[i]
		a.updateStageLocked(p.Stage)
		if p.Stage != stage || needsDiaper(*p) != (i < 3) {
			t.Fatalf("wrong care/stage on day %d", p.Age)
		}
		p.Grooming = 0
		p.Diaper = 0
		if needsMet(*p) || hygieneLevel(*p) != 0 {
			t.Fatal("active neglected care ignored")
		}
		a.generateStatusList(a.viewsList["status"])
		a.generateHygieneList(a.viewsList["hygiene"])
		expected, absent, kind, invalid := "Brushing", "Diaper", "brush", "diaper"
		if i < 3 {
			expected, absent, kind, invalid = "Diaper", "Brushing", "diaper", "brush"
		}
		for _, list := range []string{"status", "hygiene"} {
			text := ""
			for row := 0; row < a.viewsList[list].GetItemCount(); row++ {
				main, _ := a.viewsList[list].GetItemText(row)
				text += main + "\n"
			}
			if !strings.Contains(text, expected+": ") || strings.Contains(text, absent+": ") {
				t.Fatalf("%s wrong bars at %s", list, stage)
			}
		}
		a.startCare(invalid, 0)
		if a.activity != nil {
			t.Fatal("wrong age action allowed")
		}
		a.startCare(kind, 0)
		if a.activity == nil {
			t.Fatalf("active care missing at %s", stage)
		}
		a.finishCare()
		if ageCareLevel(*p) != 100 || !needsMet(*p) || hygieneLevel(*p) != 100 {
			t.Fatal("active care not restored, or inactive care still counts")
		}
	}
}

func TestDiaperToBrushingTransition(t *testing.T) {
	a := testApp(t)
	p := a.currentTamagotchi
	p.Stage = "baby"
	p.Created = time.Now().Add(-3 * 24 * time.Hour)
	p.AgeTicks = 3 * ageTicksPerDay
	p.Diaper = 0
	p.Grooming = 0
	p.GroomTicks = 20
	p.Hunger = 0
	a.startCare("diaper", 0)
	a.advanceTime(a.updateInterval)
	if p.Stage != "toddler" || p.Health != 100 || p.Grooming != 100 || p.GroomTicks != 1 {
		t.Fatal("transition retained obsolete care debt")
	}
	a.finishCare()
	if p.Diaper != 0 || p.LastChange.IsZero() == false {
		t.Fatal("stale diaper action committed after growth")
	}
	a.closeActivity()
	a.advanceTime(19 * a.updateInterval)
	if p.Grooming != 100 || p.GroomTicks != 20 {
		t.Fatal("brushing decreased too soon")
	}
	a.advanceTime(a.updateInterval)
	if p.Grooming != 99 {
		t.Fatal("brushing did not use slow decay")
	}
}

func TestBrushingDecaysSlowlyAndPersists(t *testing.T) {
	a := testApp(t)
	a.currentTamagotchi.Stage = "adult"
	a.currentTamagotchi.Created = time.Now().Add(-14 * 24 * time.Hour)
	a.currentTamagotchi.AgeTicks = 14 * ageTicksPerDay
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

func TestAgeSpecificCarePersistsWithoutResettingPet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	a := testApp(t)
	a.Config = cfg
	p := a.currentTamagotchi
	p.Name = "Saved name"
	p.Stage = "baby"
	p.Created = time.Now().Add(-2 * 24 * time.Hour)
	p.AgeTicks = 2 * ageTicksPerDay
	p.Species = "rabbit"
	p.Diaper = 37
	p.Grooming = 0
	p.Health = 70
	a.updateConfigFromState()
	cfg.App.LastLogin = time.Now()
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	loaded.App.CurrentLogin = loaded.App.LastLogin
	b := testApp(t)
	b.Config = loaded
	b.initializeStateFromConfig()
	pet := b.currentTamagotchi
	if pet.Name != "Saved name" || pet.Species != "rabbit" || pet.Health != 70 || pet.Diaper != 37 || ageCareLabel(*pet) != "Diaper" {
		t.Fatal("reload changed saved pet or active care")
	}
	pet.Created = time.Now().Add(-3 * 24 * time.Hour)
	pet.AgeTicks = 3 * ageTicksPerDay
	b.advanceTime(b.updateInterval)
	if ageCareLabel(*pet) != "Brushing" || pet.Grooming != 100 || pet.Diaper != 37 || pet.Name != "Saved name" || pet.Health != 70 {
		t.Fatal("growth reset the pet or retained obsolete care")
	}
}

// Ignore widths, spacing, coat decoration and facial features: anatomical
// growth must still change the outline, rather than stretching one template.
func TestGrowthChangesAnatomyBeyondWidth(t *testing.T) {
	for _, species := range petSpecies {
		silhouettes := map[string]bool{}
		for _, stage := range growthStages[1:] {
			pet := *newDefaultTamagotchi("Test")
			pet.Species = species
			pet.Stage = stage
			rows := actionPose(pet, "idle", 0)
			var outline []string
			for _, row := range rows {
				shape := strings.Map(func(r rune) rune {
					if strings.ContainsRune(`/\|()`, r) {
						return r
					}
					return -1
				}, row)
				if shape != "" {
					outline = append(outline, shape)
				}
			}
			silhouettes[strings.Join(outline, "\n")] = true
		}
		if len(silhouettes) < 7 {
			t.Errorf("%s: only %d anatomical growth shapes", species, len(silhouettes))
		}
	}
}

func TestTearsAreIntermittentAndPreserveSceneGeometry(t *testing.T) {
	for _, species := range petSpecies {
		for _, stage := range growthStages {
			for _, happiness := range []int{0, 10, 30, 100} {
				pet := *newDefaultTamagotchi("Test")
				pet.Species = species
				pet.Stage = stage
				pet.Happiness = happiness
				wet, dry := 0, 0
				for frame := 0; frame < 24; frame++ {
					pose := strings.Join(actionPose(pet, "idle", frame), "\n")
					if strings.Contains(pose, "💧") {
						wet++
					} else {
						dry++
					}
					scene := renderActionSprite(pet, "idle", 0, time.Duration(frame)*250*time.Millisecond)
					for _, line := range strings.Split(scene, "\n") {
						if uniseg.StringWidth(line) != actionSceneSize(pet).width {
							t.Fatal("tear shifted terminal geometry")
						}
					}
				}
				if happiness < 20 && (wet == 0 || dry == 0) {
					t.Fatalf("%s %s: no intermittent tears", species, stage)
				}
				if happiness >= 20 && wet > 0 {
					t.Fatalf("%s %s: content pet is crying", species, stage)
				}
				pet.IsAlive = false
				pet.Health = 0
				if strings.Contains(strings.Join(actionPose(pet, "idle", 5), "\n"), "💧") {
					t.Fatal("dead pet animated")
				}
			}
		}
	}
}

func TestEverySpeciesKeepsItsAgeFrameDuringMotion(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, stage := range growthStages {
			for _, species := range petSpecies {
				pet := *newDefaultTamagotchi("Frame")
				pet.Stage = stage
				pet.Species = species
				pet.compactArt = compact
				pet.Happiness = 0
				size := petArtSize(pet)
				for _, kind := range []string{"idle", "refuse", "food", "wash", "brush", "diaper", "sleep"} {
					for _, frame := range []int{0, 1, 5, 11, 16} {
						rows := actionPose(pet, kind, frame)
						if len(rows) != size.height {
							t.Fatalf("%s %s %s: height %d, want %d", species, stage, kind, len(rows), size.height)
						}
						for _, line := range rows {
							if uniseg.StringWidth(line) != size.width {
								t.Fatalf("%s %s %s: width %d, want %d", species, stage, kind, uniseg.StringWidth(line), size.width)
							}
						}
					}
				}
			}
		}
	}
}
