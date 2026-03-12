package ui

import "image/color"

// Theme defines the colour palette used to render the UI.
type Theme struct {
	// Background is the primary screen background colour.
	Background color.RGBA
	// PrimaryAccent is used for highlights, active selection, and key UI elements.
	PrimaryAccent color.RGBA
	// SecondaryAccent is used for secondary highlights and decorative elements.
	SecondaryAccent color.RGBA
	// Text is the default text colour.
	Text color.RGBA
}

// DefaultTheme is the built-in Spectacle theme.
var DefaultTheme = Theme{
	Background:      color.RGBA{0x0F, 0x05, 0x1D, 0xFF}, // Obsidian Purple
	PrimaryAccent:   color.RGBA{0xFF, 0x00, 0x7A, 0xFF}, // Electric Rose
	SecondaryAccent: color.RGBA{0x00, 0xF5, 0xFF, 0xFF}, // Cyber Cyan
	Text:            color.RGBA{0xE0, 0xE0, 0xFF, 0xFF}, // Off-white / blue tint
}
