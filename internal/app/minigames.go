package app

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

func extraKind(kind string) bool {
	switch kind {
	case "runner", "notes", "drawing", "reading", "seek":
		return true
	}
	return false
}
func (a *App) beginExtraGame(now time.Time) bool {
	g := a.activity
	if !extraKind(g.kind) {
		return false
	}
	g.phase = "playing"
	g.started = now
	g.step = 0
	g.input = 0
	g.deadline = now.Add(12 * time.Second)
	switch g.kind {
	case "runner":
		g.lane = 1
		g.sequence = make([]int, 6)
		for i := range g.sequence {
			g.sequence[i] = rand.IntN(3)
		}
		g.deadline = now.Add(5 * time.Second)
	case "notes":
		g.sequence = make([]int, 3)
		for i := range g.sequence {
			g.sequence[i] = rand.IntN(3)
		}
	case "drawing":
		g.posX, g.posY = 0, 0
		g.sequence = []int{4, 14, 10, 0}
	case "reading":
		words := []string{a.tr("Book"), a.tr("Cat"), a.tr("Sun"), a.tr("Moon"), a.tr("Tree"), a.tr("Fish")}
		g.word = a.tr(words[rand.IntN(len(words))])
		g.target = rand.IntN(3)
		first := string(unicode.ToUpper([]rune(g.word)[0]))
		g.choices = make([]string, 3)
		for i := range g.choices {
			letter := string(rune('A' + rand.IntN(26)))
			for letter == first || (i > 0 && letter == g.choices[0]) {
				letter = string(rune('A' + rand.IntN(26)))
			}
			g.choices[i] = letter
		}
		g.choices[g.target] = first
	case "seek":
		g.target = rand.IntN(3)
		g.phase = "reveal"
		g.sequence = make([]int, 3+g.round)
		for i := range g.sequence {
			g.sequence[i] = rand.IntN(2)
		}
	}
	return true
}
func (a *App) tickExtraGame(now time.Time) bool {
	g := a.activity
	if !extraKind(g.kind) {
		return false
	}
	if g.phase == "between" {
		if !now.Before(g.deadline) {
			a.beginRound(now)
		}
		return true
	}
	if g.phase == "result" || g.phase == "ready" {
		return true
	}
	if g.kind == "seek" {
		if g.phase == "reveal" && now.Sub(g.started) >= 1200*time.Millisecond {
			g.phase = "shuffle"
			g.started = now
		}
		if g.phase == "shuffle" {
			count := min(len(g.sequence), int(now.Sub(g.started)/(800*time.Millisecond)))
			for g.step < count {
				left := g.sequence[g.step]
				if g.target == left {
					g.target = left + 1
				} else if g.target == left+1 {
					g.target = left
				}
				g.step++
			}
			if g.step == len(g.sequence) {
				g.phase = "playing"
				g.deadline = now.Add(6 * time.Second)
			}
		}
	}
	if g.kind == "runner" && g.phase == "playing" {
		count := min(len(g.sequence), int(now.Sub(g.started)/(700*time.Millisecond)))
		for g.step < count {
			if g.lane == g.sequence[g.step] {
				a.finishRound(false, now)
				return true
			}
			g.step++
		}
		if g.step == len(g.sequence) {
			a.finishRound(true, now)
			return true
		}
	}
	if g.phase == "playing" && !now.Before(g.deadline) {
		a.finishRound(false, now)
	}
	return true
}
func arrowDirection(event *tcell.EventKey) int {
	switch event.Key() {
	case tcell.KeyUp:
		return 0
	case tcell.KeyRight:
		return 1
	case tcell.KeyDown:
		return 2
	case tcell.KeyLeft:
		return 3
	}
	switch unicode.ToLower(event.Rune()) {
	case 'w':
		return 0
	case 'd':
		return 1
	case 's':
		return 2
	case 'a':
		return 3
	}
	return -1
}
func (a *App) extraGameInput(event *tcell.EventKey, now time.Time) bool {
	g := a.activity
	if !extraKind(g.kind) {
		return false
	}
	switch g.kind {
	case "runner":
		direction := arrowDirection(event)
		if direction == 3 {
			g.lane = max(0, g.lane-1)
		} else if direction == 1 {
			g.lane = min(2, g.lane+1)
		}
	case "notes", "reading", "seek":
		choice := int(event.Rune() - '1')
		if choice < 0 || choice > 2 {
			return true
		}
		if g.kind == "notes" {
			if choice != g.sequence[g.input] {
				a.finishRound(false, now)
			} else {
				g.input++
				if g.input == len(g.sequence) {
					a.finishRound(true, now)
				}
			}
		} else {
			a.finishRound(choice == g.target, now)
		}
	case "drawing":
		direction := arrowDirection(event)
		if direction < 0 {
			return true
		}
		dx, dy := []int{0, 1, 0, -1}, []int{-1, 0, 1, 0}
		x, y := g.posX+dx[direction], g.posY+dy[direction]
		if x < 0 || x > 4 || y < 0 || y > 2 {
			return true
		}
		if x > 0 && x < 4 && y == 1 {
			a.finishRound(false, now)
			return true
		}
		g.posX, g.posY = x, y
		if y*5+x == g.sequence[g.input] {
			g.input++
			if g.input == len(g.sequence) {
				a.finishRound(true, now)
			}
		}
	}
	return true
}
func (a *App) extraGameText(now time.Time) string {
	g := a.activity
	if g.phase == "ready" {
		switch g.kind {
		case "runner":
			return a.tr("Dodge six obstacles. Move left/right with arrows or A/D.")
		case "notes":
			return a.tr("Sing each highlighted note. Press 1, 2 or 3.")
		case "drawing":
			return a.tr("Trace the outline clockwise using arrows or WASD.")
		case "reading":
			return a.tr("Read the word and choose its first letter with 1, 2 or 3.")
		case "seek":
			return a.tr("Watch the pet hide, follow the swaps, then choose box 1–3.")
		}
	}
	if g.phase == "between" || g.phase == "result" {
		return ""
	}
	switch g.kind {
	case "runner":
		if g.step >= len(g.sequence) {
			return ""
		}
		lanes := []string{"  ·  ", "  ·  ", "  ·  "}
		lanes[g.sequence[g.step]] = "  X  "
		feet := []string{"     ", "     ", "     "}
		feet[g.lane] = "  ^  "
		return a.tr("Dodge! ← / → or A / D") + "\n" + strings.Join(lanes, "│") + "\n" + strings.Join(feet, "│") + fmt.Sprintf(a.tr("\nObstacles passed: %d/6"), g.step)
	case "notes":
		if g.input >= len(g.sequence) {
			return ""
		}
		return fmt.Sprintf(a.tr("Sing note %d! (keys 1–3)"), g.sequence[g.input]+1) + fmt.Sprintf(a.tr("\nNotes sung: %d/3"), g.input)
	case "drawing":
		var grid strings.Builder
		for y := 0; y < 3; y++ {
			for x := 0; x < 5; x++ {
				cell := "· "
				if y*5+x == g.sequence[g.input] {
					cell = "◎ "
				}
				if x == g.posX && y == g.posY {
					cell = "✎ "
				}
				grid.WriteString(cell)
			}
			grid.WriteByte('\n')
		}
		return a.tr("Trace clockwise: top right → bottom right → bottom left → start.") + "\n" + grid.String()
	case "reading":
		return a.tr("Which is the first letter?") + "\n" + g.word + "\n1: " + g.choices[0] + "   2: " + g.choices[1] + "   3: " + g.choices[2]
	case "seek":
		if g.phase == "reveal" {
			boxes := []string{"[1]", "[2]", "[3]"}
			boxes[g.target] = "[♥]"
			return a.tr("Watch where your pet hides!") + "\n" + strings.Join(boxes, "   ")
		}
		if g.phase == "shuffle" {
			swap := min(g.step, len(g.sequence)-1)
			left := g.sequence[swap] + 1
			return fmt.Sprintf(a.tr("Swap %d/%d"), g.step+1, len(g.sequence)) + "\n" + fmt.Sprintf(a.tr("Boxes %d ↔ %d swap places!"), left, left+1) + "\n[1]   [2]   [3]"
		}
		return a.tr("Where is your pet? Press 1, 2 or 3.") + "\n[1]   [2]   [3]"
	}
	return ""
}
