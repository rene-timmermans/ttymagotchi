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
	lines                                 []string
	faceRow, mouthRow, mouthCol, waistRow int
}

// Templates use anatomical anchors so paws and held care tools follow the pet.
// Each species has nine progressively larger forms after the shared mystery egg.
func animalModel(t Tamagotchi) petModel {
	level := max(1, growthIndex(t.Stage))
	width := 22 + level
	patterns := []string{".", "..", ":.", ":.:", ". : .", ": . :", ". : : .", ": . . :", ". : . : ."}
	coat := patterns[level-1]
	span := strings.Repeat(" ", 3+level/3)
	rows := []string{}
	faceRow, mouthRow, waistRow := 0, 0, 0
	add := func(s string) {
		visible := strings.NewReplacer("{e}", "o", "{m}", "w", "{b}", "^").Replace(s)
		pad := max(0, (width-uniseg.StringWidth(visible))/2)
		rows = append(rows, strings.Repeat(" ", pad)+s)
	}
	switch t.Species {
	case "cat":
		add(`/\` + strings.Repeat(" ", 7+level/2) + `/\`)
		add(`/  \` + strings.Repeat("_", 3+level/2) + `/  \`)
		add(`/` + strings.Repeat(" ", 13+level/2) + `\`)
		faceRow = len(rows)
		add(`|  {e}` + span + `{e}  |`)
		mouthRow = len(rows)
		add(`==\` + span + `{m}` + span + `/==`)
		add(`\` + strings.Repeat("_", 9+level/2) + `/`)
		add(`/|` + span + coat + span + `|\`)
		waistRow = len(rows)
		add(`( |` + span + coat + span + `| )`)
		for i := 0; i < level/4; i++ {
			add(`| |` + strings.Repeat(" ", 11+level/2) + `| |`)
		}
		add(`\_|` + strings.Repeat("_", 11+level/2) + `|_/~~)`)
		add(`(____)` + span + `(____)`)
	case "dog":
		add(`__` + strings.Repeat("_", 9+level/2) + `__`)
		add(`/  \` + strings.Repeat("_", 5+level/2) + `/  \`)
		add(`/    |` + strings.Repeat(" ", 7+level/2) + `|    \`)
		faceRow = len(rows)
		add(`U  / {e}` + span + `{e} \  U`)
		add(`|  |` + span + `Y` + span + `|  |`)
		mouthRow = len(rows)
		add(`\__/` + span + `{m}` + span + `\__/`)
		add(`/|` + span + coat + span + `|\`)
		waistRow = len(rows)
		add(`( |` + span + coat + span + `| )`)
		for i := 0; i < level/4; i++ {
			add(`| |` + strings.Repeat(" ", 11+level/2) + `| |`)
		}
		add(`\_|` + strings.Repeat("_", 11+level/2) + `|_/ /)`)
		add(`(____)` + span + `(____)`)
	case "rabbit":
		add(`(\  /)` + strings.Repeat(" ", 3+level/3) + `(\  /)`)
		add(`| || |` + strings.Repeat(" ", 3+level/3) + `| || |`)
		if level >= 5 {
			add(`| || |` + strings.Repeat(" ", 3+level/3) + `| || |`)
		}
		add(`\_||_/` + strings.Repeat("_", 3+level/3) + `\_||_/`)
		faceRow = len(rows)
		add(`/  {e}` + span + `{e}  \`)
		mouthRow = len(rows)
		add(`==(` + span + `{m}` + span + `)==`)
		add(`\` + span + `""` + span + `/`)
		add(`/|` + span + coat + span + `|\`)
		waistRow = len(rows)
		add(`( |` + span + coat + span + `| )o`)
		for i := 0; i < level/5; i++ {
			add(`| |` + strings.Repeat(" ", 11+level/2) + `| |`)
		}
		add(`\_|` + strings.Repeat("_", 11+level/2) + `|_/`)
		add(`(______)` + span + `(______)`)
	case "guinea_pig":
		add(`__` + strings.Repeat(" ", 11+level/2) + `__`)
		add(`(  )` + strings.Repeat("_", 7+level/2) + `(  )`)
		add(`/` + strings.Repeat(" ", 15+level/2) + `\`)
		faceRow = len(rows)
		add(`|   {e}` + span + `{e}   |`)
		mouthRow = len(rows)
		add(`==(` + span + `>{m}<` + span + `)==`)
		add(`\` + strings.Repeat(" ", 15+level/2) + `/`)
		add(`/` + span + coat + span + `\`)
		waistRow = len(rows)
		add(`( ` + span + coat + span + ` )`)
		for i := 0; i < level/4; i++ {
			add(`( ` + span + coat + span + ` )`)
		}
		add(`\` + strings.Repeat("_", 17+level/2) + `/`)
		add(`(_)(_)` + span + `(_)(_)`)
	case "penguin":
		add(`.` + strings.Repeat("-", 9+level/2) + `.`)
		add(`/` + strings.Repeat(" ", 11+level/2) + `\`)
		faceRow = len(rows)
		add(`/  ({e})` + span + `({e})  \`)
		mouthRow = len(rows)
		add(`|` + span + `<{m}>` + span + `|`)
		add(`|` + span + `___` + span + `|`)
		add(`/|` + span + `/   \` + span + `|\`)
		waistRow = len(rows)
		add(`/ |` + span + `| ` + coat + ` |` + span + `| \`)
		for i := 0; i < level/4; i++ {
			add(`| |` + span + `|     |` + span + `| |`)
		}
		add(`\ |` + span + `\___/` + span + `| /`)
		add(`\_|` + strings.Repeat("_", 13+level/2) + `|_/`)
		add(`/____\` + span + `/____\`)
	}
	// Eyebrows get their own row so emotion remains readable with closed eyes.
	face := rows[faceRow]
	left := strings.Index(face, "{e}")
	right := strings.Index(strings.Replace(face, "{e}", "o", 1), "{e}")
	brows := strings.Repeat(" ", left) + "{b}" + strings.Repeat(" ", right-left-1) + "{b}"
	rows = append(rows, "")
	copy(rows[faceRow+1:], rows[faceRow:])
	rows[faceRow] = brows
	faceRow++
	mouthRow++
	waistRow++

	// Anchor columns are measured with the same one-cell eye/mouth placeholders.
	mouthLine := strings.ReplaceAll(rows[mouthRow], "{e}", "o")
	mouthCol := strings.Index(mouthLine, "{m}")
	return petModel{rows, faceRow, mouthRow, mouthCol, waistRow}
}

func animalPose(t Tamagotchi, kind string, frame int) []string {
	model := animalModel(t)
	emotion := expressionFor(t.Happiness, t.IsAlive && t.Health > 0)
	eyes, mouth := emotion.eyes, emotion.mouth
	if t.IsAlive && (kind == "sleep" || kind == "wash" || (kind == "idle" && frame%12 == 11)) {
		eyes = "-"
	}
	if t.IsAlive && kind == "food" {
		mouth = "o"
		if frame%2 == 1 {
			mouth = "-"
		}
	}
	if t.IsAlive && kind == "sleep" && frame%2 == 0 {
		mouth = "o"
	}

	rows := make([]string, len(model.lines))
	for i, line := range model.lines {
		line = strings.Replace(line, "{b}", emotion.brow, 1)
		rows[i] = strings.NewReplacer("{e}", eyes, "{m}", mouth, "{b}", emotion.rightBrow()).Replace(line)
	}
	if t.IsAlive && (kind == "wash" || kind == "brush" || kind == "diaper" || kind == "food") {
		rows = heldCarePose(rows, model, kind, frame)
	}
	if t.IsAlive && kind == "idle" {
		if t.Species == "rabbit" && frame%2 == 1 {
			rows[0] = " " + rows[0]
		}
		if t.Species == "guinea_pig" && frame%2 == 1 {
			rows[len(rows)-1] = strings.ReplaceAll(rows[len(rows)-1], "(_)(_)", "(__)()")
		}
		if t.Species == "cat" && frame%2 == 1 {
			for i := range rows {
				rows[i] = strings.ReplaceAll(rows[i], "~~)", "~(~")
			}
		}
		if t.Species == "dog" && frame%2 == 1 {
			for i := range rows {
				rows[i] = strings.ReplaceAll(rows[i], "/)", "\\)")
			}
		}
		if t.Species == "penguin" && frame%2 == 1 {
			for i := range rows {
				rows[i] = " " + rows[i]
			}
		}
	}
	if kind == "refuse" && frame%2 == 1 {
		for i := 0; i <= model.mouthRow; i++ {
			rows[i] = "  " + rows[i]
		}
	}
	return normalizePose(rows)
}

// Hands stay connected to the torso. Tools move within the body's outline;
// there are no floating water marks, sparkles or detached brushes.
func heldCarePose(rows []string, model petModel, kind string, frame int) []string {
	row := model.waistRow
	if (kind == "wash" || kind == "brush") && frame%4 >= 2 {
		row = min(row+1, len(rows)-2)
	}
	line := []rune(rows[row])
	left := len(line) - len([]rune(strings.TrimLeft(rows[row], " ")))
	right := len([]rune(strings.TrimRight(rows[row], " "))) - 1
	if right-left < 9 {
		return rows
	}
	tool := "(==)"
	if frame%2 == 1 {
		tool = "(~~)"
	}
	switch kind {
	case "brush":
		tool = "[|||]"
		if frame%2 == 1 {
			tool = "[///]"
		}
	case "diaper":
		if frame < 5 {
			tool = "[###]"
		} else if frame < 10 {
			tool = "[===]"
		} else {
			tool = "[ v ]"
		}
	case "food":
		if frame%2 == 0 {
			tool = "( o )"
		} else {
			tool = "(   )"
		}
	}
	// Reach left/right in four poses while both forelimbs remain attached.
	shift := []int{-2, 0, 2, 0}[frame%4]
	center := (left+right)/2 + shift
	start := max(left+2, min(right-len(tool)-1, center-len(tool)/2))
	for col := left + 1; col < right; col++ {
		line[col] = ' '
	}
	line[left], line[right] = '(', ')'
	line[left+1], line[right-1] = '\\', '/'
	for col := left + 2; col < start; col++ {
		line[col] = '_'
	}
	for i, c := range tool {
		line[start+i] = c
	}
	for col := start + len(tool); col < right-1; col++ {
		line[col] = '_'
	}
	rows[row] = string(line)
	return rows
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
	return normalizePose(rows)
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
