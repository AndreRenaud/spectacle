package ui

import (
	"fmt"
	"image/color"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AndreRenaud/spectacle/mpv"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var audioExtensions = map[string]bool{
	".mp3":  true,
	".flac": true,
	".ogg":  true,
	".wav":  true,
	".m4a":  true,
	".aac":  true,
	".opus": true,
	".wma":  true,
}

func isAudioFile(name string) bool {
	return audioExtensions[strings.ToLower(filepath.Ext(name))]
}

// AudioPlayer is a persistent audio player backed by MPV.
type AudioPlayer struct {
	player      *mpv.MPVPlayer
	mu          sync.Mutex
	currentSong string
	position    time.Duration
	total       time.Duration
	theme       *Theme
}

// SetTheme applies a theme for rendering.
func (ap *AudioPlayer) SetTheme(t *Theme) {
	ap.theme = t
}

// NewAudioPlayer creates a new always-active audio player.
func NewAudioPlayer() (*AudioPlayer, error) {
	ap := &AudioPlayer{}
	player, err := mpv.NewMPVPlayer(ap.mpvCallback)
	if err != nil {
		return nil, err
	}
	ap.player = player
	player.MonitorPosition(ap.positionUpdate)
	return ap, nil
}

func (ap *AudioPlayer) positionUpdate(upto, total time.Duration) {
	ap.mu.Lock()
	ap.position = upto
	ap.total = total
	ap.mu.Unlock()
}

func (ap *AudioPlayer) mpvCallback(info string) {
	log.Printf("MPV callback: %s", info)
}

// Play starts playing the given file and tracks its display title.
func (ap *AudioPlayer) Play(filePath string) error {
	base := filepath.Base(filePath)
	title := strings.TrimSuffix(base, filepath.Ext(base))
	if err := ap.player.LoadFile(filePath, false); err != nil {
		return err
	}
	ap.mu.Lock()
	ap.currentSong = title
	ap.mu.Unlock()
	return nil
}

// CurrentSong returns the title of the currently playing song, or "" if none.
func (ap *AudioPlayer) CurrentSong() string {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return ap.currentSong
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	m := total / 60
	s := total % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

const (
	audioPlayerWidth  = 420
	audioPlayerHeight = 160
	audioBarHeight    = 12
	audioPadding      = 8
)

// Draw renders the audio player widget with song title, time, and progress bar.
// x, y is the top-left corner of the widget.
func (ap *AudioPlayer) Draw(screen *ebiten.Image, x, y int) {
	ap.mu.Lock()
	song := ap.currentSong
	position := ap.position
	total := ap.total
	ap.mu.Unlock()

	if song == "" {
		return
	}

	var textColor, accentColor, barBg color.RGBA
	if ap.theme != nil {
		textColor = ap.theme.Text
		accentColor = ap.theme.SecondaryAccent
		barBg = color.RGBA{ap.theme.Text.R, ap.theme.Text.G, ap.theme.Text.B, 40}
	} else {
		textColor = color.RGBA{0xE0, 0xE0, 0xFF, 0xFF}
		accentColor = color.RGBA{0x00, 0xF5, 0xFF, 0xFF}
		barBg = color.RGBA{0xE0, 0xE0, 0xFF, 0x28}
	}

	fx, fy := float64(x), float64(y)

	// CD player icon above the song name
	if icon := icons["App_CDPlayer.png"]; icon != nil {
		iconSize := float64(48)
		iw := float64(icon.Bounds().Dx())
		ih := float64(icon.Bounds().Dy())
		scale := iconSize / max(iw, ih)
		iop := &ebiten.DrawImageOptions{}
		iop.GeoM.Scale(scale, scale)
		iop.GeoM.Translate(fx, fy)
		screen.DrawImage(icon, iop)
		fy += ih*scale + float64(audioPadding)
	}

	// Song title
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(fx, fy)
	titleOp.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, song, labelFont, titleOp)

	// Time text
	timeStr := formatDuration(position) + " / " + formatDuration(total)
	_, titleH := text.Measure(song, labelFont, 0)
	timeY := fy + titleH + float64(audioBarHeight)/2
	timeOp := &text.DrawOptions{}
	timeOp.GeoM.Translate(fx, timeY)
	timeOp.ColorScale.ScaleWithColor(accentColor)
	text.Draw(screen, timeStr, labelFont, timeOp)

	// Progress bar
	_, timeH := text.Measure(timeStr, labelFont, 0)
	barY := float32(timeY + timeH + audioBarHeight/2)
	vector.FillRect(screen, float32(x), barY, float32(audioPlayerWidth), float32(audioBarHeight), barBg, false)

	var ratio float32
	if total > 0 {
		ratio = float32(position) / float32(total)
	}
	if ratio > 1 {
		ratio = 1
	}
	if ratio > 0 {
		vector.FillRect(screen, float32(x), barY, float32(audioPlayerWidth)*ratio, float32(audioBarHeight), accentColor, false)
	}
}

// Close shuts down the underlying MPV instance.
func (ap *AudioPlayer) Close() error {
	return ap.player.Close()
}
