package app

import (
	"strings"
	"time"

	"github.com/rivo/uniseg"
)

const sceneWidth, sceneHeight = 68, 18

// A fixed-size scene preserves the pet's position while its body animates.
// Grapheme widths keep emoji food and the egg aligned with terminal cells.
type spriteScene [sceneHeight][sceneWidth]string

func newSpriteScene() *spriteScene {
	s := &spriteScene{}
	for y := range s {
		for x := range s[y] {
			s[y][x] = " "
		}
	}
	return s
}
func (s *spriteScene) put(x, y int, text string) {
	if y < 0 || y >= sceneHeight {
		return
	}
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		cell := g.Str()
		width := uniseg.StringWidth(cell)
		if width < 1 {
			continue
		}
		if x >= 0 && x+width <= sceneWidth {
			for i := 0; i < width; i++ {
				s.clearCell(x+i, y)
			}
			s[y][x] = cell
			for i := 1; i < width; i++ {
				s[y][x+i] = ""
			}
		}
		x += width
	}
}

// Erasing a wide glyph must also erase its continuation cell.
func (s *spriteScene) clearCell(x, y int) {
	start := x
	for start > 0 && s[y][start] == "" {
		start--
	}
	width := uniseg.StringWidth(s[y][start])
	for i := start; i < min(sceneWidth, start+max(1, width)); i++ {
		s[y][i] = " "
	}
}

func mouthColumn(stage string) int {
	switch stage {
	case "egg", "":
		return 2
	case "baby":
		return 3
	case "child":
		return 4
	case "teen":
		return 5
	default:
		return 6
	}
}

func (s *spriteScene) String() string {
	rows := make([]string, sceneHeight)
	for y := range s {
		rows[y] = strings.Join(s[y][:], "")
	}
	return strings.Join(rows, "\n")
}

// Change the actual face and silhouette, keeping each life stage recognizable.
func actionPose(t Tamagotchi, kind string, frame int) []string {
	if growthIndex(t.Stage) == 0 {
		return eggPose(t, kind, frame)
	}
	if knownSpecies(t.Species) {
		return animalPose(t, kind, frame)
	}
	lines := strings.Split(strings.Trim(renderTamagotchiSprite(t), "\n"), "\n")
	if !t.IsAlive {
		return lines
	}
	closed := kind == "sleep" || kind == "wash" || frame%5 == 4
	eyes := "o"
	if closed {
		eyes = "-"
	}
	mouth := "o"
	if frame%2 == 1 {
		mouth = "-"
	}
	if kind == "wash" && frame%2 == 0 {
		mouth = "v"
	}
	switch t.Stage {
	case "egg", "":
		lines[1] = " /" + eyes + eyes + "\\"
		lines[2] = " \\" + mouth + "_/"
	case "baby":
		lines[1] = " _(" + eyes + eyes + ")_"
		lines[2] = "(  " + mouth + "  )"
	case "child":
		lines[1] = " ( " + eyes + " " + eyes + " )"
		lines[2] = " /  " + mouth + "  \\"
	case "teen":
		lines[1] = "  ( " + eyes + " " + eyes + " )"
		lines[2] = "  /| " + mouth + " |\\"
	default:
		lines[1] = "  ( " + eyes + "   " + eyes + " )"
		lines[2] = "  /|  " + mouth + "  |\\"
	}
	if kind == "sleep" && frame%2 == 0 {
		// A fuller chest on inhalation; the next pose exhales and settles.
		switch t.Stage {
		case "baby":
			lines[2] = "(   o   )"
		case "child":
			lines[2] = " / (o) \\"
		case "teen":
			lines[2] = " / | o | \\"
		case "adult":
			lines[2] = " / |  o  | \\"
		}
	}
	if kind == "brush" || kind == "diaper" {
		// The pet shifts its feet while its coat is brushed.
		last := len(lines) - 1
		if frame%2 == 0 {
			lines[last] = "  _/   \\_"
		} else {
			lines[last] = "   \\___/"
		}
	}
	return lines
}

func renderActionSprite(t Tamagotchi, kind string, index int, elapsed time.Duration) string {
	if !t.IsAlive {
		return renderTamagotchiSprite(t)
	}
	frame := int(maxDuration(0, elapsed) / (250 * time.Millisecond))
	pose := actionPose(t, kind, frame)
	s := newSpriteScene()
	width := 0
	for _, line := range pose {
		width = max(width, uniseg.StringWidth(line))
	}
	x, y := (sceneWidth-width)/2, 1
	mouthRow, mouthCol := 6, 11
	if growthIndex(t.Stage) > 0 && knownSpecies(t.Species) {
		model := animalModel(t)
		mouthRow, mouthCol = model.mouthRow, model.mouthCol
	}
	for row, line := range pose {
		s.put(x, y+row, line)
	}
	switch kind {
	case "food":
		// Hold the selected food at the mouth instead of flying it across the scene.
		if elapsed < 1500*time.Millisecond && frame%3 != 2 {
			icon := strings.Fields(availableFoods[index].Name)[0]
			s.put(x+mouthCol, y+mouthRow, icon)
		}
	case "sleep":
		// Keep the requested Zzz close to the head; breathing comes from the body.
		if frame%2 == 0 {
			s.put(x+width-5, y, "Zzz")
		}
	}
	return s.String()
}
