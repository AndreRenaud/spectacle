package ui

import (
	"embed"
	_ "embed"
	"image"
	"image/color"
	_ "image/png"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	// BorderSize is the pixel width of the border around the grid
	BorderSize int

	entries     []fs.DirEntry
	theme       *Theme
	selectedIdx int // index of the currently selected entry
	cols        int // number of columns from the last Draw call
}

// NewBrowser creates a Browser rooted at dir inside fsys.
// cellSize is the pixel width/height of each icon cell.
func NewBrowser(fsys fs.ReadDirFS, dir string, cellSize int, borderSize int, labelFont *text.GoTextFace) *Browser {
	b := &Browser{
		FS:         fsys,
		Dir:        dir,
		CellSize:   cellSize,
		BorderSize: borderSize,
		Font:       labelFont,
	}
	b.reload()
	return b
}

type dirEntry struct {
	name  string
	isDir bool
}

func (d *dirEntry) Name() string {
	return d.name
}

func (d *dirEntry) IsDir() bool {
	return d.isDir
}

func (d *dirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}
func (d *dirEntry) Type() fs.FileMode {
	if d.isDir {
		return fs.ModeDir
	}
	return 0
}

func (b *Browser) reload() {
	entries, err := b.FS.ReadDir(b.Dir)
	if err != nil {
		log.Printf("browser: ReadDir %q: %v", b.Dir, err)
		b.entries = nil
		return
	}
	if b.Dir != "." {
		// Prepend a synthetic ".." entry for navigating up to the parent directory.
		entries = append([]fs.DirEntry{&dirEntry{name: "..", isDir: true}}, entries...)
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
	b.cols = (screenW - x) / b.CellSize
	if b.cols < 1 {
		b.cols = 1
	}

	cellHeight := b.CellSize
	if b.Font != nil {
		cellHeight += int(b.Font.Size) + 4
	}

	for i, entry := range b.entries {
		col := i % b.cols
		row := i / b.cols

		cx := x + col*b.CellSize
		cy := y + row*cellHeight

		// Draw selection highlight behind the icon.
		if i == b.selectedIdx && b.theme != nil {
			ac := b.theme.PrimaryAccent
			highlight := color.RGBA{ac.R, ac.G, ac.B, 80}
			vector.FillRect(screen, float32(cx), float32(cy), float32(b.CellSize), float32(cellHeight), highlight, false)
		}

		icon := icons["File_Generic.png"]
		name := entry.Name()
		if name == ".." {
			icon = icons["Action_GoBack_2.png"]
			name = "Parent Directory"
		} else if entry.IsDir() {
			icon = icons["Folder_generic.png"]
		}

		iconW := float64(icon.Bounds().Dx())
		iconH := float64(icon.Bounds().Dy())
		scale := float64(b.CellSize-b.BorderSize*2) / max(iconW, iconH)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(cx+b.BorderSize), float64(cy+b.BorderSize))
		screen.DrawImage(icon, op)

		if b.Font != nil {
			labelOp := &text.DrawOptions{}
			labelOp.GeoM.Translate(float64(cx+b.BorderSize), float64(cy+b.BorderSize)+iconH*scale)
			if b.theme != nil {
				labelOp.ColorScale.ScaleWithColor(b.theme.Text)
			}
			text.Draw(screen, name, b.Font, labelOp)
		}
	}
}

// MoveSelection moves the selected icon by (dx, dy) cells.
// For example, MoveSelection(-1, 0) moves one cell to the left,
// MoveSelection(0, 1) moves one row down.
func (b *Browser) MoveSelection(dx, dy int) {
	if len(b.entries) == 0 {
		return
	}
	cols := b.cols
	if cols < 1 {
		cols = 1
	}
	newIdx := b.selectedIdx + dx + dy*cols
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(b.entries) {
		newIdx = len(b.entries) - 1
	}
	b.selectedIdx = newIdx
}

// SelectItem activates the currently selected entry.
// If it is a directory, the browser navigates into it and returns nil.
// If it is a file, the full path within the FS is returned.
func (b *Browser) SelectItem() (string, bool) {
	if len(b.entries) == 0 || b.selectedIdx >= len(b.entries) {
		return "", false
	}
	entry := b.entries[b.selectedIdx]
	if entry.IsDir() {
		// Navigate into the directory.
		newDir := b.Dir + "/" + entry.Name()
		if b.Dir == "." {
			newDir = entry.Name()
		}
		b.Dir = filepath.Clean(newDir)
		b.selectedIdx = 0
		b.reload()
		return "", false
	}
	// Return the full path for files.
	path := b.Dir + "/" + entry.Name()
	if b.Dir == "." {
		path = entry.Name()
	}
	return filepath.Clean(path), true
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
