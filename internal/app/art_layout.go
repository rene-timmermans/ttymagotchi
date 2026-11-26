package app

import (
	"github.com/rivo/uniseg"
	"strings"
)

type artSize struct{ width, height int }

// Shared by every species. A viewport chooses one age-sized frame, never a
// different scale for an individual animal. The baseline is shared as well.
var fullArtSizes = []artSize{{24, 10}, {26, 12}, {30, 15}, {34, 18}, {38, 21}, {42, 24}, {46, 27}, {50, 30}, {54, 32}, {58, 34}}
var compactArtSizes = []artSize{{24, 10}, {24, 10}, {26, 11}, {28, 12}, {30, 13}, {32, 14}, {34, 15}, {34, 16}, {34, 16}, {34, 16}}

func petArtSize(p Tamagotchi) artSize {
	index := growthIndex(p.Stage)
	size := fullArtSizes[index]
	if p.compactArt {
		size = compactArtSizes[index]
	}
	if p.artHeightLimit > 0 {
		size.height = min(size.height, p.artHeightLimit)
	}
	return size
}
func petForViewport(p Tamagotchi, width, height, reserved int) Tamagotchi {
	size := fullArtSizes[len(fullArtSizes)-1]
	if width <= 0 {
		width = 38
	}
	if height <= 0 {
		height = 22
	}
	p.compactArt = width < size.width || height < size.height+reserved
	return p
}
func poseFrame(rows []string, size artSize) []string {
	width := 0
	for _, line := range rows {
		width = max(width, uniseg.StringWidth(line))
	}
	left := max(0, (size.width-width)/2)
	top := max(0, size.height-len(rows))
	framed := make([]string, 0, max(size.height, len(rows)))
	for i := 0; i < top; i++ {
		framed = append(framed, strings.Repeat(" ", size.width))
	}
	for _, line := range rows {
		line = strings.Repeat(" ", left) + line
		framed = append(framed, line+strings.Repeat(" ", max(0, size.width-uniseg.StringWidth(line))))
	}
	return framed
}
func frameTemplate(rows []string, size artSize) []string {
	visible := make([]string, len(rows))
	replacer := strings.NewReplacer("{b}", "^", "{e}", "o", "{m}", "w", "{c}", " ", "{p}", "_", "{t}", ")", "{a}", ")")
	width := 0
	for i, row := range rows {
		visible[i] = replacer.Replace(row)
		width = max(width, uniseg.StringWidth(visible[i]))
	}
	left := max(0, (size.width-width)/2)
	out := make([]string, 0, max(size.height, len(rows)))
	for i := 0; i < max(0, size.height-len(rows)); i++ {
		out = append(out, "")
	}
	for _, row := range rows {
		out = append(out, strings.Repeat(" ", left)+row+strings.Repeat(" ", max(0, size.width-left-uniseg.StringWidth(replacer.Replace(row)))))
	}
	return out
}
func actionSceneSize(p Tamagotchi) artSize {
	size := petArtSize(p)
	return artSize{min(sceneWidth, size.width+6), min(sceneHeight, size.height+2)}
}

// Motion may use the existing empty margins, but must not change the age frame.
func padPoseInFrame(rows []string, size artSize) []string {
	out := make([]string, 0, max(len(rows), size.height))
	for _, line := range rows {
		if uniseg.StringWidth(line) > size.width {
			line = strings.TrimRight(line, " ")
		}
		out = append(out, line+strings.Repeat(" ", max(0, size.width-uniseg.StringWidth(line))))
	}
	for len(out) < size.height {
		out = append(out, strings.Repeat(" ", size.width))
	}
	return out
}
