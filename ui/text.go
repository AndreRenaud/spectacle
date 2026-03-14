package ui

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// DrawWrapped draws str onto dst, wrapping lines so that no line exceeds
// maxWidth pixels. Each line is drawn using the same DrawOptions (position,
// colour, etc.). The Y origin is advanced by one line-height between lines;
// all other DrawOptions fields are unchanged.
//
// Word wrapping uses a greedy algorithm: words are appended to the current
// line until adding the next word would exceed maxWidth, at which point a new
// line is started. If a single word is still wider than maxWidth it is broken
// at the character level to fit.
func DrawWrapped(dst *ebiten.Image, str string, face *text.GoTextFace, maxWidth int, alignment Alignment, opts *text.DrawOptions) int {
	if str == "" {
		return 0
	}

	// Measure the line-height once so every row advances by the same amount.
	_, lineH := text.Measure(" ", face, opts.LineSpacing)

	// Clone the options so we can mutate the Y translation per line without
	// altering the caller's value.
	o := *opts

	// drawLine renders a single pre-broken line, applying horizontal alignment.
	drawLine := func(line string) {
		lo := o
		if alignment != AlignLeft {
			lw, _ := text.Measure(line, face, opts.LineSpacing)
			offset := float64(maxWidth) - lw
			if alignment == AlignCenter {
				offset /= 2
			}
			lo.GeoM.Translate(offset, 0)
		}
		text.Draw(dst, line, face, &lo)
	}

	words := strings.Fields(str)
	lineStart := 0
	totalHeight := 0

	for lineStart < len(words) {
		// Build the longest line that still fits within maxWidth.
		lineEnd := lineStart
		for lineEnd < len(words) {
			candidate := strings.Join(words[lineStart:lineEnd+1], " ")
			w, _ := text.Measure(candidate, face, opts.LineSpacing)
			if int(w) > maxWidth && lineEnd > lineStart {
				// This word pushes us over the limit; stop before it.
				break
			}
			lineEnd++
		}

		line := strings.Join(words[lineStart:lineEnd], " ")

		// If even a single word exceeds maxWidth, break it at the character level.
		w, _ := text.Measure(line, face, opts.LineSpacing)
		if int(w) > maxWidth {
			runes := []rune(line)
			charStart := 0
			for charStart < len(runes) {
				charEnd := charStart + 1
				for charEnd < len(runes) {
					candidate := string(runes[charStart : charEnd+1])
					cw, _ := text.Measure(candidate, face, opts.LineSpacing)
					if int(cw) > maxWidth {
						break
					}
					charEnd++
				}
				drawLine(string(runes[charStart:charEnd]))
				o.GeoM.Translate(0, lineH)
				totalHeight += int(lineH)
				charStart = charEnd
			}
		} else {
			drawLine(line)
			o.GeoM.Translate(0, lineH)
			totalHeight += int(lineH)
		}

		lineStart = lineEnd
	}
	return totalHeight
}
