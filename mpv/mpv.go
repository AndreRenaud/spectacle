package mpv

// See https://mpv.io/manual/stable/ for details on the API

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gen2brain/go-mpv"
)

type MPVPlayer struct {
	speed        float64
	m            *mpv.Mpv
	positionFunc func(upto, total time.Duration)
	title        string
}

// NewMPVPlayer starts a player, and it will call the `callback` function whenever
// we start a new chapter (passing the chapter title into the function)
// TODO: If we have an audio path latency, can we compensate with the `audio-delay` parameter?
func NewMPVPlayer() (*MPVPlayer, error) {
	m := mpv.New()

	if err := m.RequestLogMessages("info"); err != nil {
		return nil, err
	}
	//if err := m.ObserveProperty(0, "pause", mpv.FormatFlag); err != nil {
	//return nil, err
	//}
	//if err := m.ObserveProperty(1, "volume", mpv.FormatInt64); err != nil {
	//return nil, err
	//}
	if err := m.ObserveProperty(0, "chapter", mpv.FormatInt64); err != nil {
		return nil, err
	}
	if err := m.ObserveProperty(0, "time-pos/full", mpv.FormatDouble); err != nil {
		return nil, err
	}
	if err := m.ObserveProperty(0, "duration/full", mpv.FormatDouble); err != nil {
		return nil, err
	}
	//if err := m.ObserveProperty(0, "media-title", mpv.FormatString); err != nil {
	//return nil, err
	//}

	//_ = m.SetPropertyString("input-default-bindings", "yes")
	//_ = m.SetOptionString("input-vo-keyboard", "yes")
	//_ = m.SetOption("osc", mpv.FormatFlag, true)

	// Turn on low-latency audio, as per `mpv --show-profile=low-latency`
	if err := m.SetProperty("audio-buffer", mpv.FormatInt64, 0); err != nil {
		return nil, err
	}
	if err := m.SetProperty("vd-lavc-threads", mpv.FormatInt64, 1); err != nil {
		return nil, err
	}
	if err := m.SetPropertyString("cache-pause", "no"); err != nil {
		return nil, fmt.Errorf("cache pause: %w", err)
	}
	//if err := m.SetPropertyString("demuxer-lavf-o-add", "fflags=+nobuffer"); err != nil {
	//return nil, fmt.Errorf("demuxer-lavf-o-add: %w", err)
	//}
	if err := m.SetPropertyString("demuxer-lavf-probe-info", "nostreams"); err != nil {
		return nil, fmt.Errorf("demuxer-lavf-probe-info: %w", err)
	}
	if err := m.SetProperty("demuxer-lavf-analyzeduration", mpv.FormatDouble, 0.1); err != nil {
		return nil, fmt.Errorf("demuxer-lavf-analyzeduration: %w", err)
	}
	if err := m.SetPropertyString("video-sync", "audio"); err != nil {
		return nil, fmt.Errorf("video-sync: %w", err)
	}
	if err := m.SetPropertyString("interpolation", "no"); err != nil {
		return nil, fmt.Errorf("interpolation: %w", err)
	}
	if err := m.SetPropertyString("video-latency-hacks", "yes"); err != nil {
		return nil, fmt.Errorf("video-latency-hacks: %w", err)
	}
	if err := m.SetPropertyString("stream-buffer-size", "4k"); err != nil {
		return nil, fmt.Errorf("stream-buffer-size: %w", err)
	}
	if err := m.SetPropertyString("af", "@level:lavfi=\"astats=metadata=1:reset=4\""); err != nil {
		return nil, fmt.Errorf("af: %w", err)
	}

	if err := m.Initialize(); err != nil {
		m.TerminateDestroy()
		return nil, err
	}

	retval := &MPVPlayer{m: m, speed: 1.0}
	go retval.run()
	return retval, nil
}

func (m *MPVPlayer) GetTitle() string {
	return m.title
}

func (m *MPVPlayer) LoadFile(filename string, loop bool) error {
	// We don't get told directly from MPV about missing files, so manually check first
	// The MPV errors come back through the event system in m.run
	if _, err := os.Stat(filename); err != nil {
		return err
	}
	if loop {
		if err := m.m.SetOptionString("loop-file", "inf"); err != nil {
			return fmt.Errorf("loop-file: %w", err)
		}
	} else {
		if err := m.m.SetOptionString("loop-file", "no"); err != nil {
			return fmt.Errorf("loop-file: %w", err)
		}
	}

	if err := m.m.Command([]string{"loadfile", filename}); err != nil {
		return err
	}
	m.m.SetProperty("playlist-pos", mpv.FormatInt64, 0)

	return nil
}

func (m *MPVPlayer) MonitorPosition(mon func(upto, total time.Duration)) {
	m.positionFunc = mon
}

func (m *MPVPlayer) Stop() error {
	m.m.Command([]string{"stop"})
	m.m.Command([]string{"playlist-clear"})
	return nil
}

func (m *MPVPlayer) SetSpeed(speed float64) error {
	log.Printf("Setting speed to %f", speed)
	m.speed = speed
	dir := "forward"
	if speed < 0 {
		dir = "backward"
		speed = -speed
	}

	if err := m.m.SetPropertyString("play-direction", dir); err != nil {
		return err
	}

	return m.m.SetProperty("speed", mpv.FormatDouble, speed)
}

func (m *MPVPlayer) Speed() float64 {
	return m.speed
}

func (m *MPVPlayer) run() {
	var currentDuration time.Duration
	for {
		e := m.m.WaitEvent(10000)

		switch e.EventID {
		case mpv.EventPropertyChange:
			prop := e.Property()

			switch prop.Name {
			case "duration/full":
				if prop.Data != nil {
					duration := time.Duration(prop.Data.(float64)*float64(time.Second)) * time.Nanosecond
					log.Printf("Duration: %s", duration)
					currentDuration = duration
					if m.positionFunc != nil {
						m.positionFunc(0, currentDuration)
					}
				}
			case "time-pos/full":
				if prop.Data != nil {
					position := time.Duration(prop.Data.(float64)*float64(time.Second)) * time.Nanosecond
					//log.Printf("Position: %s", position)
					if m.positionFunc != nil {
						m.positionFunc(position, currentDuration)
					}
				}
			//case "media-title":
			//log.Printf("property: %s data=%#v format=%d", prop.Name, prop.Data, prop.Format)
			//if prop.Data != nil {
			//title := prop.Data.(string)
			//log.Printf("Title: %s", title)
			//m.title = title
			//}
			default:
				log.Printf("property: %s data=%#v format=%d", prop.Name, prop.Data, prop.Format)
			}
		case mpv.EventFileLoaded:
			if p, err := m.m.GetProperty("media-title", mpv.FormatString); err != nil {
				log.Printf("Load error: %s", err)
			} else {
				m.title = p.(string)
				log.Printf("title: %s", p.(string))
			}
		case mpv.EventLogMsg:
			log.Printf("message: %s", e.LogMessage().Text)
		case mpv.EventStart:
			sf := e.StartFile()
			log.Printf("start: %d", sf.EntryID)
		case mpv.EventEnd:
			ef := e.EndFile()
			log.Printf("end: %d %s", ef.EntryID, ef.Reason)
			if ef.Reason == mpv.EndFileError {
				log.Printf("error: %s", ef.Error)
			}
			m.title = ""
		case mpv.EventShutdown:
			log.Printf("shutdown: %d", e.EventID)
			return

		default:
			log.Printf("event: %s[%d]", e.EventID, e.EventID)
		}

		if e.Error != nil {
			log.Printf("mpv event error: %s", e.Error)
		}
	}
}

func (m *MPVPlayer) Close() error {
	m.m.TerminateDestroy()
	return nil
}

func (m *MPVPlayer) Pause(state bool) error {
	return m.m.SetProperty("pause", mpv.FormatFlag, state)
}

func (m *MPVPlayer) Paused() bool {
	val, err := m.m.GetProperty("pause", mpv.FormatFlag)
	if err != nil {
		log.Printf("cannot get pause property: %s", err)
		return false
	}
	return val.(bool)
}

func (m *MPVPlayer) AudioLevel() (float64, error) {
	val, err := m.m.GetProperty("af-metadata/level", mpv.FormatString)
	if err != nil {
		return 0, err
	}
	vals := map[string]string{}
	if err := json.Unmarshal([]byte(val.(string)), &vals); err != nil {
		return 0, fmt.Errorf("unmarshal: %w", err)
	}
	//log.Printf("vals: %#v", vals)
	if level, ok := vals["lavfi.astats.Overall.Max_level"]; ok {
		return strconv.ParseFloat(level, 64)
	}
	return 0, fmt.Errorf("not found")
}
