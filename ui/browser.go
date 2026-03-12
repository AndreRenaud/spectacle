package ui

import (
	"embed"
	_ "embed"
	"image"
	_ "image/png"
	"io/fs"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed resources/icons/*.png
var iconFiles embed.FS

var (
	icons map[string]*ebiten.Image
)

func init() {
	icons = make(map[string]*ebiten.Image)
	entries, err := iconFiles.ReadDir("resources/icons")
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		f, err := iconFiles.Open("resources/icons/" + entry.Name())
		if err != nil {
			log.Printf("failed to open icon %q: %v", entry.Name(), err)
			continue
		}
		img, _, err := image.Decode(f)
		if err != nil {
			log.Printf("failed to decode icon %q: %v", entry.Name(), err)
			continue
		}
		icons[entry.Name()] = ebiten.NewImageFromImage(img)
	}
}

// Browser displays the contents of an fs.ReadDirFS as a grid of icons.
type Browser struct {
	// FS is the filesystem to read entries from.
	FS fs.ReadDirFS
	// Dir is the current directory path within FS (use "." for the root).
	Dir string
	// CellSize is the width and height of each grid cell in logical pixels.
	CellSize int
	// Font is used to render file names
	Font *text.GoTextFace

	entries []fs.DirEntry
	theme   *Theme
}

// NewBrowser creates a Browser rooted at dir inside fsys.
// cellSize is the pixel width/height of each icon cell.
func NewBrowser(fsys fs.ReadDirFS, dir string, cellSize int, labelFont *text.GoTextFace) *Browser {
	b := &Browser{
		FS:       fsys,
		Dir:      dir,
		CellSize: cellSize,
		Font:     labelFont,
	}
	b.reload()
	return b
}

func (b *Browser) reload() {
	entries, err := b.FS.ReadDir(b.Dir)
	if err != nil {
		log.Printf("browser: ReadDir %q: %v", b.Dir, err)
		b.entries = nil
		return
	}
	b.entries = entries
}

// SetTheme applies a theme to the browser (used by the engine on Enter).
func (b *Browser) SetTheme(t *Theme) {
	b.theme = t
}

// Update handles input for the browser. Returns nil (no navigation change for now).
func (b *Browser) Update() {
}

// Draw renders the file grid onto screen, starting at pixel offset (x, y).
func (b *Browser) Draw(screen *ebiten.Image, x, y int) {
	if len(b.entries) == 0 {
		return
	}

	screenW := screen.Bounds().Dx()
	cols := (screenW - x) / b.CellSize
	if cols < 1 {
		cols = 1
	}

	cellHeight := b.CellSize
	if b.Font != nil {
		cellHeight += int(b.Font.Size) + 4
	}

	for i, entry := range b.entries {
		col := i % cols
		row := i / cols

		cx := x + col*b.CellSize
		cy := y + row*cellHeight

		icon := icons["File_Generic.png"]
		if entry.IsDir() {
			icon = icons["Folder_generic.png"]
		}

		iconW := float64(icon.Bounds().Dx())
		iconH := float64(icon.Bounds().Dy())
		scale := float64(b.CellSize) / max(iconW, iconH)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(cx), float64(cy))
		screen.DrawImage(icon, op)

		if b.Font != nil {
			labelOp := &text.DrawOptions{}
			labelOp.GeoM.Translate(float64(cx), float64(cy+b.CellSize+4))
			if b.theme != nil {
				labelOp.ColorScale.ScaleWithColor(b.theme.Text)
			}
			text.Draw(screen, entry.Name(), b.Font, labelOp)
		}
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
