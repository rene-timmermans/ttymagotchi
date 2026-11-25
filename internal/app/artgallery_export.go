//go:build artgallery

package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rene-timmermans/ttymagotchi/internal/config"
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"
)

type galleryRun struct {
	Text  string `json:"text"`
	Cells int    `json:"cells"`
}
type galleryCard struct {
	Section string         `json:"section"`
	Species string         `json:"species"`
	Stage   string         `json:"stage"`
	Label   string         `json:"label"`
	Frames  []int          `json:"frames,omitempty"`
	Text    string         `json:"text"`
	Lines   [][]galleryRun `json:"lines"`
}
type galleryData struct {
	FrameMilliseconds int           `json:"frame_ms"`
	Species           []string      `json:"species"`
	Stages            []string      `json:"stages"`
	Cards             []galleryCard `json:"cards"`
}

func galleryLines(text string) [][]galleryRun {
	result := [][]galleryRun{}
	for _, line := range strings.Split(text, "\n") {
		runs := []galleryRun{}
		ascii := strings.Builder{}
		flush := func() {
			if ascii.Len() > 0 {
				s := ascii.String()
				runs = append(runs, galleryRun{s, len(s)})
				ascii.Reset()
			}
		}
		graphemes := uniseg.NewGraphemes(line)
		for graphemes.Next() {
			s := graphemes.Str()
			if len(s) == 1 && s[0] < 128 {
				ascii.WriteString(s)
			} else {
				flush()
				runs = append(runs, galleryRun{s, uniseg.StringWidth(s)})
			}
		}
		flush()
		result = append(result, runs)
	}
	return result
}

// ExportArtGallery is build-tagged tooling: it exports the actual game renderers
// without starting the simulation or opening/saving any player configuration.
func ExportArtGallery(path string) error {
	data := galleryData{FrameMilliseconds: 250, Species: petSpecies, Stages: growthStages}
	add := func(section, species, stage, label, text string, frames []int) {
		data.Cards = append(data.Cards, galleryCard{section, species, stage, label, frames, text, galleryLines(text)})
	}
	frames := func(section, species, stage, label string, count int, render func(int) string) {
		seen := map[string]int{}
		for frame := 0; frame < count; frame++ {
			text := render(frame)
			if index, ok := seen[text]; ok {
				data.Cards[index].Frames = append(data.Cards[index].Frames, frame)
			} else {
				seen[text] = len(data.Cards)
				add(section, species, stage, label, text, []int{frame})
			}
		}
	}
	for _, species := range petSpecies {
		for _, stage := range growthStages {
			pet := *newDefaultTamagotchi("Gallery")
			pet.Species = species
			pet.Stage = stage
			pet.Age = growthDays[growthIndex(stage)]
			pet.Happiness = 50
			add("Growth", species, stage, "Neutral growth form", renderTamagotchiSprite(pet), nil)
			for _, mood := range []struct {
				name      string
				happiness int
			}{{"Delighted", 100}, {"Content", 70}, {"Neutral", 50}, {"Worried", 30}, {"Sad", 10}, {"Tearful", 0}, {"Dead", 0}} {
				p := pet
				p.Happiness = mood.happiness
				if mood.name == "Dead" {
					p.IsAlive = false
					p.Health = 0
				}
				add("Expressions", species, stage, mood.name, renderTamagotchiSprite(p), nil)
			}
			frames("Idle", species, stage, "Idle (400 ms per frame)", 12, func(frame int) string {
				if stage != "egg" {
					return strings.Join(animalPose(pet, "idle", frame), "\n")
				}
				pose := eggPose(pet, "idle", frame)
				if frame%2 == 1 {
					for i := range pose {
						pose[i] = " " + pose[i]
					}
				}
				return strings.Join(pose, "\n")
			})
			for _, animation := range []struct {
				kind, label string
				count       int
			}{{"food", "Eating", 12}, {"sleep", "Sleeping", 16}, {"wash", "Washing", 16}, {ageCareKind(pet), ageCareLabel(pet), 16}, {"refuse", "Refusal (sleep / full stomach)", 12}} {
				frames("Care", species, stage, animation.label, animation.count, func(frame int) string {
					return renderActionSprite(pet, animation.kind, 0, time.Duration(frame)*250*time.Millisecond)
				})
			}
			// Every selectable food has its own approaching sprite.
			for food := 1; food < len(availableFoods); food++ {
				add("Food", species, stage, availableFoods[food].Name, renderActionSprite(pet, "food", food, 750*time.Millisecond), []int{3})
			}
		}
	}
	// Retained upstream ASCII remains visible for review, clearly marked as legacy.
	for _, legacy := range []struct {
		stage   string
		sprites map[string]string
	}{{"egg", eggSprites}, {"baby", babySprites}, {"child", childSprites}, {"teen", teenSprites}, {"adult", adultSprites}} {
		for _, mood := range []string{"neutral", "happy", "sad", "dead"} {
			add("Legacy", "upstream", legacy.stage, mood, strings.Trim(legacy.sprites[mood], "\n"), nil)
		}
	}
	at := time.Unix(1700000000, 0)
	for index, game := range availableGames {
		pet := newDefaultTamagotchi("Gallery")
		pet.Stage = "adult"
		pet.Species = petSpecies[index%len(petSpecies)]
		a := &App{TApp: tview.NewApplication(), Config: &config.Config{App: config.AppConfig{Language: "en"}}, viewsList: make(map[string]*tview.List), currentTamagotchi: pet}
		a.playWithTamagotchi(index)
		a.activity.started = at
		a.renderActivity(at)
		add("Games", "interface", "ready", game.Name, a.activity.view.GetText(true), nil)
		a.beginRound(at)
		g := a.activity
		switch g.kind {
		case "runner":
			g.sequence = []int{0, 2, 1, 0, 2, 1}
			g.lane = 1
		case "notes":
			g.sequence = []int{2, 0, 1}
		case "memory":
			g.sequence = []int{0, 1, 3}
		case "arrows":
			g.sequence = []int{0, 1, 2, 3, 0}
		case "reading":
			g.word = "Book"
			g.choices = []string{"B", "C", "D"}
			g.target = 0
		case "seek":
			g.target = 1
			g.sequence = []int{0, 1, 0}
		}
		a.renderActivity(at.Add(500 * time.Millisecond))
		add("Games", "interface", "playing", game.Name, a.activity.view.GetText(true), nil)
		if g.kind == "memory" || g.kind == "seek" {
			g.phase = "playing"
			a.renderActivity(at.Add(2 * time.Second))
			add("Games", "interface", "answer", game.Name, a.activity.view.GetText(true), nil)
		}
	}
	for _, scenario := range []struct {
		name, kind             string
		hunger, health, energy int
	}{{"Sleep refusal", "sleep", 50, 100, 80}, {"Full stomach: stop selected", "food", 0, 100, 80}, {"Full stomach: continue selected", "food", 0, 100, 80}, {"Full stomach: fatal warning", "food", 0, 5, 80}} {
		p := newDefaultTamagotchi("Gallery")
		p.Stage = "adult"
		p.Species = "cat"
		p.Hunger = scenario.hunger
		p.Health = scenario.health
		p.Energy = scenario.energy
		a := &App{TApp: tview.NewApplication(), Config: &config.Config{App: config.AppConfig{Language: "en"}}, viewsList: make(map[string]*tview.List), currentTamagotchi: p}
		if scenario.kind == "sleep" {
			a.startCare("sleep", 0)
		} else {
			a.startFeeding(0)
			a.activity.phase = "choice"
			if strings.Contains(scenario.name, "continue") {
				a.activity.input = 1
			}
		}
		a.activity.started = at
		a.renderActivity(at.Add(time.Second))
		add("Dialogs", "interface", "choice", scenario.name, a.activity.view.GetText(true), nil)
	}
	output, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, output, 0644); err != nil {
		return err
	}
	fmt.Printf("Exported %d distinct labeled art cards to %s\n", len(data.Cards), path)
	return nil
}
