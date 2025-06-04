// src/player/player.go

package player

import (
	"fmt"
	"log"
	"sync"
	"time"

	vlc "github.com/adrg/libvlc-go/v3"
)

// Config holds player configuration
type Config struct {
	DefaultVolume    float64
	HardwareAccel    bool
	SubtitleFontSize int
	CacheDir         string
}

// Player is a media player using libVLC
type Player struct {
	config *Config
	player *vlc.Player
	mutex  sync.Mutex

	// State
	isPlaying bool
	isPaused  bool
	isStopped bool
	duration  time.Duration
	position  time.Duration

	// Event callbacks
	onStateChanged    func(string)
	onTimeChanged     func(time.Duration)
	onPositionChanged func(float32)
	onEndReached      func()
	onError           func(error)
}

// NewPlayer creates a new media player
func NewPlayer(config *Config) (*Player, error) {
	// Create a new player directly using vlc.NewPlayer()
	player, err := vlc.NewPlayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create media player: %w", err)
	}

	// Set initial volume
	if err := player.SetVolume(int(config.DefaultVolume * 100)); err != nil {
		player.Release()
		return nil, fmt.Errorf("failed to set volume: %w", err)
	}

	// Create player instance
	p := &Player{
		config:    config,
		player:    player,
		isPlaying: false,
		isPaused:  false,
		isStopped: true,
		duration:  0,
		position:  0,
	}

	// Register event handlers
	eventManager, err := player.EventManager()
	if err != nil {
		player.Release()
		return nil, fmt.Errorf("failed to get event manager: %w", err)
	}

	// Register for events you're interested in
	eventID, err := eventManager.Attach(vlc.MediaPlayerPlaying, p.handlePlaying, nil)
	if err != nil {
		log.Printf("Failed to attach playing event: %v", err)
	}

	eventID, err = eventManager.Attach(vlc.MediaPlayerPaused, p.handlePaused, nil)
	if err != nil {
		log.Printf("Failed to attach paused event: %v", err)
	}

	eventID, err = eventManager.Attach(vlc.MediaPlayerStopped, p.handleStopped, nil)
	if err != nil {
		log.Printf("Failed to attach stopped event: %v", err)
	}

	eventID, err = eventManager.Attach(vlc.MediaPlayerEndReached, p.handleEndReached, nil)
	if err != nil {
		log.Printf("Failed to attach end reached event: %v", err)
	}

	eventID, err = eventManager.Attach(vlc.MediaPlayerTimeChanged, p.handleTimeChanged, nil)
	if err != nil {
		log.Printf("Failed to attach time changed event: %v", err)
	}

	eventID, err = eventManager.Attach(vlc.MediaPlayerPositionChanged, p.handlePositionChanged, nil)
	if err != nil {
		log.Printf("Failed to attach position changed event: %v", err)
	}

	// Prevent unused variable warning
	_ = eventID

	return p, nil
}

// Event handler methods
func (p *Player) handlePlaying(event vlc.Event, userData interface{}) {
	p.mutex.Lock()
	p.isPlaying = true
	p.isPaused = false
	p.isStopped = false
	callback := p.onStateChanged
	p.mutex.Unlock()

	if callback != nil {
		callback("playing")
	}
}

func (p *Player) handlePaused(event vlc.Event, userData interface{}) {
	p.mutex.Lock()
	p.isPlaying = true
	p.isPaused = true
	p.isStopped = false
	callback := p.onStateChanged
	p.mutex.Unlock()

	if callback != nil {
		callback("paused")
	}
}

func (p *Player) handleStopped(event vlc.Event, userData interface{}) {
	p.mutex.Lock()
	p.isPlaying = false
	p.isPaused = false
	p.isStopped = true
	callback := p.onStateChanged
	p.mutex.Unlock()

	if callback != nil {
		callback("stopped")
	}
}

func (p *Player) handleEndReached(event vlc.Event, userData interface{}) {
	p.mutex.Lock()
	p.isPlaying = false
	p.isPaused = false
	p.isStopped = true
	endCallback := p.onEndReached
	stateCallback := p.onStateChanged
	p.mutex.Unlock()

	if stateCallback != nil {
		stateCallback("stopped")
	}

	if endCallback != nil {
		endCallback()
	}
}

func (p *Player) handleTimeChanged(event vlc.Event, userData interface{}) {
	p.mutex.Lock()
	callback := p.onTimeChanged
	p.mutex.Unlock()

	if callback == nil {
		return
	}

	// Get time
	t, err := p.player.MediaTime()
	if err != nil {
		return
	}

	callback(time.Duration(t) * time.Millisecond)
}

func (p *Player) handlePositionChanged(event vlc.Event, userData interface{}) {
	p.mutex.Lock()
	callback := p.onPositionChanged
	p.mutex.Unlock()

	if callback == nil {
		return
	}

	// Get position
	position, err := p.player.MediaPosition()
	if err != nil {
		return
	}

	callback(position)
}
