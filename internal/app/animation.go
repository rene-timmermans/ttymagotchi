package app

import (
	"strings"
	"time"

	"github.com/rivo/uniseg"
)

const sceneWidth, sceneHeight = 68, 18

// A fixed-size scene preserves the pet's position while effects move around it.
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
	faceRow, mouthRow, mouthCol, waistRow := 1, 2, mouthColumn(t.Stage), min(3, len(pose)-1)
	if growthIndex(t.Stage) == 0 {
		faceRow, mouthRow, mouthCol, waistRow = 5, 6, 11, 8
	} else if knownSpecies(t.Species) {
		model := animalModel(t)
		faceRow, mouthRow, mouthCol, waistRow = model.faceRow, model.mouthRow, model.mouthCol, model.waistRow
	}
	if kind == "wash" {
		x += frame % 2
	}
	if kind == "sleep" {
		y += frame % 2
	}
	for row, line := range pose {
		s.put(x, y+row, line)
	}
	faceY := y + faceRow
	mouthY := y + mouthRow
	switch kind {
	case "food":
		icon := strings.Fields(availableFoods[index].Name)[0]
		if elapsed < 1500*time.Millisecond {
			// The selected food travels straight toward the mouth in the pet scene.
			foodX := 2 + (x+mouthCol-2)*min(12, int(elapsed/(125*time.Millisecond)))/12
			s.put(foodX, mouthY, icon)
		} else if frame%2 == 0 {
			s.put(x+10, mouthY, "·")
			s.put(x+9, mouthY+1, "·")
		}
	case "sleep":
		// Zs rise from the head, alongside the breathing body.
		s.put(x+width+2, faceY, "Zzz")
		s.put(x+width+5, faceY-1-frame%2, "Z")
		s.put(x+width+7, 0, "Z")
		s.put(x-2, y+len(pose), "("+strings.Repeat("_", width+2)+")")
	case "wash":
		// Falling water, bubbles on the body and a sponge moving over the fur.
		for i := 0; i < 4; i++ {
			s.put(x-2+i*4, (frame+i)%2, "│")
		}
		s.put(x-2, faceY+frame%3, "o")
		s.put(x+width+1, faceY+(frame+1)%3, "O")
		s.put(x+mouthCol-3+frame%5, y+waistRow, "▣▣")
		water := []string{strings.Repeat("~ ", width/2+2), strings.Repeat(" ~", width/2+2)}
		s.put(x-2, y+len(pose)-1, water[frame%2])
		s.put(x-3, y+len(pose), "\\"+strings.Repeat("_", width+4)+"/")
	case "diaper":
		waistY := y + waistRow
		if frame < 5 {
			s.put(x+mouthCol-4, waistY, "\\__###__/")
		} else if frame < 10 {
			s.put(x+mouthCol-4+frame%5, waistY, "▱▱")
		} else {
			s.put(x+mouthCol-4, waistY, "\\___♡___/")
		}
		if frame%2 == 0 {
			s.put(x-2, faceY, "✧")
		} else {
			s.put(x+width+1, faceY, "✧")
		}

	case "brush":
		// The bristles touch the coat and sweep across the chest and lower body.
		brushY := y + waistRow + (frame/4)%2
		brushX := x + mouthCol - 5 + frame%8
		s.put(brushX, brushY-1, "╭────╮")
		s.put(brushX, brushY, "┴┴┴┴┴┴")
		s.put(brushX+5, brushY-1, "──")
		if frame%2 == 0 {
			s.put(x-2, faceY, "✧")
		} else {
			s.put(x+width+1, faceY, "✧")
		}

	}
	return s.String()
}
