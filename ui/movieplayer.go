package ui

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AndreRenaud/spectacle/mpv"
)

var videoExtensions = map[string]bool{
	".mp4":  true,
	".avi":  true,
	".mkv":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
}

func isVideoFile(name string) bool {
	return videoExtensions[strings.ToLower(filepath.Ext(name))]
}

// MoviePlayer plays video files using MPV, which opens its own native window.
type MoviePlayer struct {
	player   *mpv.MPVPlayer
	mu       sync.Mutex
	position time.Duration
	total    time.Duration
	theme    *Theme
}

// NewMoviePlayer creates a new always-active movie player.
func NewMoviePlayer() (*MoviePlayer, error) {
	mp := &MoviePlayer{}
	player, err := mpv.NewMPVPlayer()
	if err != nil {
		return nil, err
	}
	mp.player = player
	player.MonitorPosition(mp.positionUpdate)
	return mp, nil
}

// SetTheme applies a theme for rendering the overlay.
func (mp *MoviePlayer) SetTheme(t *Theme) {
	mp.theme = t
}

func (mp *MoviePlayer) positionUpdate(upto, total time.Duration) {
	mp.mu.Lock()
	mp.position = upto
	mp.total = total
	mp.mu.Unlock()
}

// Play starts playing the given video file. MPV opens its own window for rendering.
func (mp *MoviePlayer) Play(filePath string) error {
	if err := mp.player.LoadFile(filePath, false); err != nil {
		return err
	}
	mp.mu.Lock()
	mp.position = 0
	mp.total = 0
	mp.mu.Unlock()
	return nil
}

// Stop halts playback.
func (mp *MoviePlayer) Stop() error {
	mp.mu.Lock()
	mp.position = 0
	mp.total = 0
	mp.mu.Unlock()
	return mp.player.Stop()
}

// Close shuts down the underlying MPV instance.
func (mp *MoviePlayer) Close() error {
	return mp.player.Close()
}

func (mp *MoviePlayer) SetGeometry(x, y, width, height int) {
	mp.player.SetGeometry(x, y, width, height)
}
