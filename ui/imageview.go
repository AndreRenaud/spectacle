package ui

import (
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"path"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// imageExtensions lists the file suffixes treated as images.
var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
}

func isImageFile(name string) bool {
	return imageExtensions[strings.ToLower(path.Ext(name))]
}

const defaultTicksPerSlide = 60 * 5 // 5 seconds at 60 fps

// ImageView displays images from a directory as an automatic slideshow.
// It starts at the given initial file and cycles through all image files in
// the same directory in order, looping indefinitely until Escape is pressed.
type ImageView struct {
	fsys          fs.ReadDirFS
	dir           string   // directory within fsys containing the images
	files         []string // image filenames in dir
	currentIdx    int      // index into files of the currently displayed image
	current       *ebiten.Image
	tick          int // frames elapsed on the current slide
	TicksPerSlide int // frames per slide; defaults to defaultTicksPerSlide
}

// NewImageView creates an ImageView rooted in fsys, starting at initialFile.
// initialFile is a path relative to fsys (e.g. "photos/beach.jpg").
// All image files in the same directory are included in the slideshow.
func NewImageView(fsys fs.ReadDirFS, initialFile string) (*ImageView, error) {
	dir := path.Dir(initialFile)
	if dir == "" {
		dir = "."
	}

	entries, err := fsys.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && isImageFile(e.Name()) {
			files = append(files, e.Name())
		}
	}

	// Find the starting index.
	startIdx := 0
	base := path.Base(initialFile)
	for i, f := range files {
		if f == base {
			startIdx = i
			break
		}
	}

	iv := &ImageView{
		fsys:          fsys,
		dir:           dir,
		files:         files,
		currentIdx:    startIdx,
		TicksPerSlide: defaultTicksPerSlide,
	}
	iv.loadCurrent()
	return iv, nil
}

func (iv *ImageView) loadCurrent() {
	if len(iv.files) == 0 {
		iv.current = nil
		return
	}
	filePath := iv.files[iv.currentIdx]
	if iv.dir != "." {
		filePath = iv.dir + "/" + filePath
	}
	f, err := iv.fsys.Open(filePath)
	if err != nil {
		log.Printf("imageview: open %q: %v", filePath, err)
		iv.current = nil
		return
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Printf("imageview: decode %q: %v", filePath, err)
		iv.current = nil
		return
	}
	iv.current = ebiten.NewImageFromImage(img)
}

func (iv *ImageView) advance() {
	if len(iv.files) == 0 {
		return
	}
	iv.currentIdx = (iv.currentIdx + 1) % len(iv.files)
	iv.tick = 0
	iv.loadCurrent()
}

// Update advances the slideshow tick, loading the next image when the slide
// duration expires. Key handling is the responsibility of the caller.
func (iv *ImageView) Update() {
	if len(iv.files) == 0 {
		return
	}
	iv.tick++
	if iv.tick >= iv.TicksPerSlide {
		iv.advance()
	}
}

// Draw renders the current image scaled to fill screen, centred with
// letterboxing/pillarboxing to preserve aspect ratio.
func (iv *ImageView) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	if iv.current == nil {
		return
	}

	sw := float64(screen.Bounds().Dx())
	sh := float64(screen.Bounds().Dy())
	iw := float64(iv.current.Bounds().Dx())
	ih := float64(iv.current.Bounds().Dy())

	scale := min(sw/iw, sh/ih)
	scaledW := iw * scale
	scaledH := ih * scale

	ox := (sw - scaledW) / 2
	oy := (sh - scaledH) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(ox, oy)
	screen.DrawImage(iv.current, op)
}

// CurrentFile returns the path of the currently displayed image.
func (iv *ImageView) CurrentFile() string {
	if len(iv.files) == 0 {
		return ""
	}
	if iv.dir == "." {
		return iv.files[iv.currentIdx]
	}
	return iv.dir + "/" + iv.files[iv.currentIdx]
}
