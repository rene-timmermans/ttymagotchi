package app

import (
	"strings"
	"testing"
	"time"
)

func TestSixExpressionsAcrossAllGrowthForms(t *testing.T) {
	for _, species := range petSpecies {
		for _, stage := range growthStages {
			p := *newDefaultTamagotchi("Test")
			p.Species = species
			p.Stage = stage
			faces := map[string]bool{}
			faceRow, mouthRow := 5, 6
			if stage != "egg" {
				model := animalModel(p)
				faceRow, mouthRow = model.faceRow, model.mouthRow
			}
			for _, happiness := range []int{100, 70, 50, 30, 10, 0} {
				p.Happiness = happiness
				idle := actionPose(p, "idle", 0)
				face := idle[faceRow-1] + idle[faceRow] + idle[mouthRow]
				if faces[face] {
					t.Fatalf("%s %s duplicated Happiness expression at %d", species, stage, happiness)
				}
				faces[face] = true
				for _, kind := range []string{"food", "sleep", "wash", "brush"} {
					pose := actionPose(p, kind, 0)
					if pose[faceRow-1] != idle[faceRow-1] {
						t.Fatalf("%s %s %s hides Happiness expression", species, stage, kind)
					}
				}
			}
		}
	}
}

func TestNeglectGraduallyLowersHappiness(t *testing.T) {
	for _, neglect := range []string{"hunger", "cleanliness", "brushing", "energy"} {
		a := testApp(t)
		p := a.currentTamagotchi
		p.Hunger = 0
		p.Happiness = 100
		switch neglect {
		case "hunger":
			p.Hunger = 80
		case "cleanliness":
			p.Cleanliness = 20
		case "brushing":
			p.Grooming = 20
			p.Stage = "adult"
			p.Created = time.Now().Add(-14 * 24 * time.Hour)
		case "energy":
			p.Energy = 19
		}
		a.advanceTime(5 * a.updateInterval)
		if p.Happiness != 95 {
			t.Fatalf("%s did not reduce Happiness gradually: %d", neglect, p.Happiness)
		}
	}
}

func TestResultKeepsActualHappinessExpression(t *testing.T) {
	a := testApp(t)
	a.currentTamagotchi.Species = "cat"
	a.currentTamagotchi.Stage = "baby"
	a.currentTamagotchi.Happiness = 0
	a.startFeeding(0)
	a.tickActivity(a.activity.started.Add(4 * time.Second))
	p, _ := a.tamagotchiSnapshot()
	if p.Happiness != 5 {
		t.Fatal("wrong food Happiness")
	}
	expected := renderTamagotchiSprite(p)
	if !strings.Contains(a.activity.view.GetText(false), expected) {
		t.Fatal("result replaced real mood with a happy face")
	}
}

func TestNeglectedAgeCareHurtsHealthAtEveryAge(t *testing.T) {
	for i, stage := range growthStages {
		a := testApp(t)
		p := a.currentTamagotchi
		p.Stage = stage
		p.Created = time.Now().Add(-time.Duration(growthDays[i]) * 24 * time.Hour)
		p.Grooming = 0
		p.Diaper = 0
		p.Hunger = 0
		a.advanceTime(a.updateInterval)
		if p.Health != 95 || p.Happiness != 49 || needsMet(*p) {
			t.Fatalf("brushing neglect ignored at %s", stage)
		}
	}
}
