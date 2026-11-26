package app

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/rivo/uniseg"
)

var petSpecies = []string{"cat", "dog", "rabbit", "guinea_pig", "penguin"}
var growthStages = []string{"egg", "hatchling", "baby", "toddler", "youngster", "child", "preteen", "teen", "adolescent", "adult"}
var growthDays = []int{0, 1, 2, 3, 4, 5, 7, 9, 12, 14}

func randomSpecies() string { return petSpecies[rand.IntN(len(petSpecies))] }
func knownSpecies(species string) bool {
	for _, s := range petSpecies {
		if s == species {
			return true
		}
	}
	return false
}
func growthIndex(stage string) int {
	for i, s := range growthStages {
		if stage == s {
			return i
		}
	}
	return 0
}
func needsDiaper(t Tamagotchi) bool { return growthIndex(t.Stage) < 3 }
func ageCareLevel(t Tamagotchi) int {
	if needsDiaper(t) {
		return t.Diaper
	}
	return t.Grooming
}
func ageCareLabel(t Tamagotchi) string {
	if needsDiaper(t) {
		return "Diaper"
	}
	return "Brushing"
}
func ageCareIcon(t Tamagotchi) string {
	if needsDiaper(t) {
		return "🧷"
	}
	return "🪮"
}
func ageCareKind(t Tamagotchi) string {
	if needsDiaper(t) {
		return "diaper"
	}
	return "brush"
}
func hygieneLevel(t Tamagotchi) int { return min(t.Cleanliness, ageCareLevel(t)) }

func speciesLabel(t Tamagotchi) string {
	if growthIndex(t.Stage) == 0 {
		return "Mystery egg"
	}
	switch t.Species {
	case "cat":
		return "Cat"
	case "dog":
		return "Dog"
	case "rabbit":
		return "Rabbit"
	case "guinea_pig":
		return "Guinea pig"
	case "penguin":
		return "Penguin (Tux)"
	default:
		return "Tamagotchi"
	}
}

type petModel struct {
	lines                                                                []string
	faceRow, mouthRow, mouthCol, waistRow, careCol, eyeCol, secondEyeCol int
}

func animalModel(t Tamagotchi) petModel {
	rows := frameTemplate(portraitFor(t), petArtSize(t))
	model := petModel{lines: rows}
	for row, line := range rows {
		if strings.Contains(line, "{e}") {
			model.faceRow = row
			model.eyeCol = strings.Index(line, "{e}")
			model.secondEyeCol = strings.Index(strings.Replace(line, "{e}", "o", 1), "{e}")
		}
		if col := strings.Index(line, "{m}"); col >= 0 {
			model.mouthRow = row
			model.mouthCol = col
		}
		if col := strings.Index(line, "{c}"); col >= 0 {
			model.waistRow = row
			model.careCol = col
		}
	}
	return model
}

func animalPose(t Tamagotchi, kind string, frame int) []string {
	model := animalModel(t)
	alive := t.IsAlive && t.Health > 0
	emotion := expressionFor(t.Happiness, alive)
	eyes, mouth := emotion.eyes, emotion.mouth
	phase := frame % 24
	if alive && kind == "idle" {
		if phase == 11 || phase == 12 {
			eyes = "-"
		}
		if t.Energy < 20 && phase >= 16 && phase <= 18 {
			eyes = "-"
			mouth = "o"
		}
		if t.Species == "guinea_pig" && phase >= 3 && phase <= 5 {
			mouth = "o"
		}
	}
	if alive && (kind == "sleep" || kind == "wash" || (kind == "idle" && frame%12 == 11)) {
		eyes = "-"
	}
	if alive && (kind == "food" || kind == "sleep") {
		if frame%2 == 0 {
			mouth = "o"
		} else {
			mouth = emotion.mouth
		}
	}
	paw, tail, ear := "_", ")", ")"
	if t.Species == "dog" {
		tail = "/"
	}
	if t.Species == "rabbit" {
		tail = "o"
	}
	if alive && kind == "idle" {
		switch t.Species {
		case "cat":
			if phase >= 3 && phase <= 6 && phase%2 == 1 {
				tail = "/"
			}
		case "dog":
			if t.Happiness >= 40 && phase >= 2 && phase <= 8 && phase%2 == 1 {
				tail = "\\"
			}
		case "rabbit":
			if phase >= 4 && phase <= 6 {
				ear = "|"
			}
		case "guinea_pig":
			if phase == 4 {
				paw = "/"
			}
		case "penguin":
			if phase >= 4 && phase <= 7 {
				if phase%2 == 0 {
					paw = "/"
				} else {
					paw = "\\"
				}
			}
		}
	}

	if alive && (kind == "wash" || kind == "brush" || kind == "diaper" || kind == "food") {
		if frame%2 == 0 {
			paw = "/"
		} else {
			paw = "\\"
		}
	}
	rows := make([]string, len(model.lines))
	for i, line := range model.lines {
		line = strings.Replace(line, "{b}", emotion.brow, 1)
		rows[i] = strings.NewReplacer("{b}", emotion.rightBrow(), "{e}", eyes, "{m}", mouth, "{c}", " ", "{p}", paw, "{t}", tail, "{a}", ear).Replace(line)
	}
	if alive && kind == "sleep" {
		// The paws settle on the ground; mouth and chest move with each breath.
		if frame%2 == 1 {
			rows[model.waistRow] = replacePetCells(rows[model.waistRow], model.careCol, "(")
		}
	}
	if alive && (kind == "wash" || kind == "brush" || kind == "diaper") {
		tool := "(=)"
		switch kind {
		case "wash":
			if frame%2 == 1 {
				tool = "(~)"
			}
		case "brush":
			tool = "[|]"
			if frame%2 == 1 {
				tool = "[/]"
			}
		case "diaper":
			if frame < 5 {
				tool = "[#]"
			} else if frame < 10 {
				tool = "[=]"
			} else {
				tool = "[ ]"
			}
		}
		// Contact points are authored with each drawing, so tools touch the animal.
		rows[model.waistRow] = replacePetCells(rows[model.waistRow], model.careCol, tool)
	}
	if alive && kind == "idle" {
		rows = addPetTear(rows, model, t, phase)
		// Young animals briefly wobble on short legs; adults stay grounded.
		if growthIndex(t.Stage) <= 2 && phase >= 5 && phase <= 7 {
			for i := range rows {
				rows[i] = " " + rows[i]
			}
		}
		// A waddle moves the complete penguin, keeping its outline connected.
		if t.Species == "penguin" && phase >= 5 && phase <= 7 {
			for i := range rows {
				rows[i] = " " + rows[i]
			}
		}
	}
	if alive && kind == "refuse" && frame%2 == 1 {
		for i := 0; i <= model.mouthRow; i++ {
			rows[i] = "  " + rows[i]
		}
	}
	return padPoseInFrame(rows, petArtSize(t))
}

// A tear starts at an eye and follows a short path down the cheek. Each cycle
// has a pause; very unhappy pets shed a second tear from the other eye.
func addPetTear(rows []string, model petModel, pet Tamagotchi, phase int) []string {
	if pet.Happiness >= 20 {
		return rows
	}
	first := 7
	if pet.Happiness <= 0 {
		first = 3
	}
	step := phase - first
	column := model.eyeCol
	if pet.Happiness <= 0 && phase >= 15 {
		step = phase - 15
		column = model.secondEyeCol
	}
	if step < 0 || step >= 4 {
		return rows
	}
	row := model.faceRow + 1 + step
	if row >= len(rows) {
		return rows
	}
	if step > 1 {
		column = max(0, column-step+1)
	}
	rows[row] = replacePetCells(rows[row], column, "💧")
	return rows
}

// Use the scene's grapheme-aware cell overlay for both ASCII and wide emoji.
func replacePetCells(line string, column int, mark string) string {
	scene := newSpriteScene()
	scene.put(0, 0, line)
	scene.put(column, 0, mark)
	width := min(sceneWidth, max(uniseg.StringWidth(line), column+uniseg.StringWidth(mark)))
	return strings.Join(scene[0][:width], "")
}

func stageLabel(t Tamagotchi) string {
	return fmt.Sprintf("%s (%d/10)", t.Stage, growthIndex(t.Stage)+1)
}

func eggPose(t Tamagotchi, kind string, frame int) []string {
	emotion := expressionFor(t.Happiness, t.IsAlive && t.Health > 0)
	eyes, mouth := emotion.eyes, emotion.mouth
	if t.IsAlive && (kind == "sleep" || kind == "wash" || (kind == "idle" && frame%12 == 11)) {
		eyes = "-"
	}
	if t.IsAlive && (kind == "food" || kind == "sleep") {
		mouth = "o"
		if frame%2 == 1 {
			mouth = emotion.mouth
		}
	}

	rows := []string{
		`       .--------.`,
		`     .'          '.`,
		`    /              \`,
		`   /                \`,
		`  |     ` + emotion.brow + `    ` + emotion.rightBrow() + `       |`,
		`  |     ` + eyes + `    ` + eyes + `       |`,
		`  |        ` + mouth + `         |`,
		`   \                /`,
		`    \    .  .      /`,
		`     '------------'`,
	}
	if kind == "wash" || kind == "brush" || kind == "diaper" {
		tool := "[==]"
		if kind == "brush" {
			tool = "[|||]"
		}
		if kind == "diaper" {
			if frame < 5 {
				tool = "[###]"
			} else if frame < 10 {
				tool = "[===]"
			} else {
				tool = "[ v ]"
			}
		}
		pad := 3 + []int{0, 2, 4, 2}[frame%4]
		rows[7] = `   \` + strings.Repeat(" ", pad) + tool + strings.Repeat(" ", 16-pad-len(tool)) + `/`
	}
	if kind == "refuse" && frame%2 == 1 {
		for i := 0; i <= 6; i++ {
			rows[i] = "  " + rows[i]
		}
	}
	if t.IsAlive && t.Health > 0 && kind == "idle" {
		rows = addPetTear(rows, petModel{faceRow: 5, eyeCol: 8, secondEyeCol: 13}, t, frame%24)
	}

	return poseFrame(rows, petArtSize(t))
}

// Keep all rows equally wide so tview centers the whole animal as one figure.
func normalizePose(rows []string) []string {
	width := 0
	for _, line := range rows {
		width = max(width, uniseg.StringWidth(line))
	}
	for i, line := range rows {
		rows[i] = line + strings.Repeat(" ", width-uniseg.StringWidth(line))
	}
	return rows
}
