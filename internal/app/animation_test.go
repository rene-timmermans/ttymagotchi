package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

func TestAllBarsShowEveryPercent(t *testing.T) {
	a := testApp(t)
	previous := ""
	for percent := 0; percent <= 100; percent++ {
		bar := a.createProgressBar(percent, 100)
		fill := strings.SplitN(bar, " ", 2)[1]
		if fill == previous {
			t.Fatalf("no visible change at %d%%", percent)
		}
		if width := uniseg.StringWidth(fill); width != 20 {
			t.Fatalf("bar width %d at %d%%", width, percent)
		}
		previous = fill
	}
}

func TestPetBodyAnimatesForEveryStage(t *testing.T) {
	for _, stage := range []string{"egg", "baby", "child", "teen", "adult"} {
		for _, kind := range []string{"food", "sleep", "wash", "brush"} {
			pet := *newDefaultTamagotchi("Test")
			pet.Stage = stage
			// Compare actual anatomy, excluding captions and decorative effects.
			first := strings.Join(actionPose(pet, kind, 0), "\n")
			second := strings.Join(actionPose(pet, kind, 1), "\n")
			if first == second {
				t.Fatalf("%s %s pet body is static", stage, kind)
			}
			for frame := 0; frame < 16; frame++ {
				scene := renderActionSprite(pet, kind, 0, time.Duration(frame)*250*time.Millisecond)
				lines := strings.Split(scene, "\n")
				if len(lines) != actionSceneSize(pet).height {
					t.Fatalf("%s %s unstable scene height", stage, kind)
				}
				for _, line := range lines {
					if width := uniseg.StringWidth(line); width != actionSceneSize(pet).width {
						t.Fatalf("%s %s frame %d width=%d: %q", stage, kind, frame, width, line)
					}
				}
			}
		}
	}
}

func TestAnimationVisibleAtTerminalCells(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(80, 24)
	for _, kind := range []string{"food", "sleep", "wash", "brush"} {
		a := testApp(t)
		a.currentTamagotchi.Stage = "adult"

		if kind == "food" {
			a.startFeeding(0)
		} else {
			a.startCare(kind, 0)
		}
		view := a.activity.view
		view.SetRect(0, 0, 80, 24)
		now := a.activity.started
		a.renderActivity(now)
		view.Draw(screen)
		screen.Show()
		cells, _, _ := screen.GetContents()
		first := make([]string, len(cells))
		for i, c := range cells {
			first[i] = string(c.Runes)
		}
		a.renderActivity(now.Add(250 * time.Millisecond))
		view.Draw(screen)
		screen.Show()
		second, _, _ := screen.GetContents()
		changed := false
		// The sprite lives above the explanatory caption. Check only that region.
		for y := 2; y < 20; y++ {
			for x := 20; x < 60; x++ {
				i := y*80 + x
				if first[i] != string(second[i].Runes) {
					changed = true
				}
			}
		}
		if !changed {
			t.Fatalf("%s has no visible animation in sprite area", kind)
		}
	}
}

func TestDeadPetDoesNotAnimate(t *testing.T) {
	pet := *newDefaultTamagotchi("Test")
	pet.IsAlive = false
	pet.Health = 0
	for _, kind := range []string{"food", "sleep", "wash", "brush"} {
		if renderActionSprite(pet, kind, 0, 0) != renderTamagotchiSprite(pet) {
			t.Fatalf("dead pet animated for %s", kind)
		}
	}
}
