package ui

import (
	"bytes"
	_ "embed"
	"io/fs"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed resources/Michroma-Regular.ttf
var michromaFontData []byte

const (
	ScreenWidth  = 1920
	ScreenHeight = 1080
)

var (
	titleFont  *text.GoTextFace
	labelFont  *text.GoTextFace
	fontSource *text.GoTextFaceSource
)

func init() {
	var err error
	fontSource, err = text.NewGoTextFaceSource(bytes.NewReader(michromaFontData))
	if err != nil {
		log.Fatal(err)
	}
	titleFont = &text.GoTextFace{Source: fontSource, Size: 64}
	labelFont = &text.GoTextFace{Source: fontSource, Size: 18}
}

type Game struct {
	Theme   Theme
	browser *Browser
}

func (g *Game) Update() error {
	if g.browser != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			g.browser.MoveSelection(-1, 0)
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			g.browser.MoveSelection(1, 0)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			g.browser.MoveSelection(0, -1)
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			g.browser.MoveSelection(0, 1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if path := g.browser.SelectItem(); path != nil {
				log.Printf("selected file: %s", *path)
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(g.Theme.Background)
	op := &text.DrawOptions{}
	op.GeoM.Translate(60, 100)
	op.ColorScale.ScaleWithColor(g.Theme.Text)
	text.Draw(screen, "Spectacle Media Player", titleFont, op)
	if g.browser != nil {
		g.browser.Draw(screen, 0, 200)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	return ScreenWidth, ScreenHeight
}

// Run starts the Spectacle media player UI browsing fsys.
func Run(fsys fs.ReadDirFS) error {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Spectacle Media Player")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	game := &Game{
		Theme:   DefaultTheme,
		browser: NewBrowser(fsys, ".", 96, 8, labelFont),
	}
	game.browser.SetTheme(&game.Theme)
	return ebiten.RunGame(game)
}
