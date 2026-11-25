package app

import (
	"strings"
	"time"

	"github.com/rivo/uniseg"
)

const sceneWidth, sceneHeight = 34, 11

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
	if kind == "change" {
		// The pet kicks its feet as the diaper is removed, wiped and replaced.
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
	x, y := 12, 2
	if kind == "wash" {
		x += frame % 2
	}
	if kind == "sleep" {
		y += frame % 2
	}
	for row, line := range pose {
		s.put(x, y+row, line)
	}
	faceY := y + 1
	mouthY := y + 2
	switch kind {
	case "food":
		icon := strings.Fields(availableFoods[index].Name)[0]
		if elapsed < 1500*time.Millisecond {
			// The selected food travels straight toward the mouth in the pet scene.
			foodX := 2 + (x+mouthColumn(t.Stage)-2)*min(12, int(elapsed/(125*time.Millisecond)))/12
			s.put(foodX, mouthY, icon)
		} else if frame%2 == 0 {
			s.put(x+10, mouthY, "·")
			s.put(x+9, mouthY+1, "·")
		}
	case "sleep":
		// Zs rise from the head, alongside the breathing body.
		s.put(x+14, faceY, "Zzz")
		s.put(x+18, faceY-1-frame%2, "Z")
		s.put(x+20, 0, "Z")
		s.put(x-2, y+len(pose), "(___ soft pillow ___)")
	case "wash":
		// Falling water, bubbles on the body and a sponge moving over the fur.
		for i := 0; i < 4; i++ {
			s.put(x-2+i*4, (frame+i)%2, "│")
		}
		s.put(x-2, faceY+frame%3, "o")
		s.put(x+14, faceY+(frame+1)%3, "O")
		s.put(x+2+frame%4, mouthY, "▣")
		water := []string{"~ ~ ~ ~ ~ ~ ~ ~", " ~ ~ ~ ~ ~ ~ ~ "}
		s.put(x-2, y+len(pose)-1, water[frame%2])
		s.put(x-3, y+len(pose), "\\_______________/")
	case "change":
		waistY := y + min(3, len(pose)-1)
		if waistY < mouthY {
			waistY = mouthY
		}
		phase := min(3, frame/4)
		switch phase {
		case 0:
			s.put(x+mouthColumn(t.Stage)-2, waistY, "\\_#_/")
			s.put(x-4-frame%4, waistY, "←")
		case 1:
			s.put(x+mouthColumn(t.Stage)-2+frame%4, waistY, "▱")
			s.put(x+8, waistY, "·")
		default:
			s.put(x+mouthColumn(t.Stage)-2, waistY, "\\_♡_/")
			if frame%2 == 0 {
				s.put(x-2, faceY, "✧")
			} else {
				s.put(x+14, faceY, "✧")
			}
		}
	}
	return s.String()
}
