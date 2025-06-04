// src/app/app.go

package app

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"

	"github.com/N0tT1m/sync-stream/src/config"
	"github.com/N0tT1m/sync-stream/src/player"
	"github.com/N0tT1m/sync-stream/src/plex"
	"github.com/N0tT1m/sync-stream/src/sync"
	"github.com/N0tT1m/sync-stream/src/ui"
)

// Application holds the main application state
type Application struct {
	// App components
	fyneApp      fyne.App
	mainWindow   fyne.Window
	config       *config.Config
	plexClient   *plex.Client
	syncClient   *sync.Client
	player       *player.Player
	uiController *ui.Controller

	// State
	isAuthenticated bool
	currentMedia    *plex.MediaItem
}

// New creates a new application instance
func New(cfg *config.Config) (*Application, error) {
	// Create Fyne application
	fyneApp := app.New()

	// Apply theme from config
	applyTheme(fyneApp, cfg)

	// Create main window
	mainWindow := fyneApp.NewWindow("Plex Desktop Client")
	mainWindow.Resize(fyne.NewSize(1280, 720))
	mainWindow.SetMaster()

	// Create application instance
	app := &Application{
		fyneApp:         fyneApp,
		mainWindow:      mainWindow,
		config:          cfg,
		isAuthenticated: false,
	}

	// Initialize components
	if err := app.initializeComponents(); err != nil {
		return nil, err
	}

	return app, nil
}

// initializeComponents sets up all the application components
func (a *Application) initializeComponents() error {
	// Initialize Plex client
	a.plexClient = plex.NewClient(&plex.Config{
		ServerURL:        a.config.PlexServerURL,
		ClientIdentifier: a.config.PlexClientIdentifier,
		Username:         a.config.PlexUsername,
		Token:            a.config.PlexToken,
	})

	// Initialize sync client
	a.syncClient = sync.NewClient(&sync.Config{
		ServerURL:    a.config.SyncServerURL,
		SyncInterval: a.config.SyncInterval,
		Enabled:      a.config.SyncEnabled,
	})

	// Initialize player
	var err error
	a.player, err = player.NewPlayer(&player.Config{
		DefaultVolume:    a.config.DefaultVolume,
		HardwareAccel:    a.config.EnableHardwareAccel,
		SubtitleFontSize: a.config.SubtitleFontSize,
		CacheDir:         a.config.CacheDir,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize player: %w", err)
	}

	// Initialize UI controller
	a.uiController = ui.NewController(a.fyneApp, a.mainWindow, a.player)

	// Setup interoperability between components
	a.setupInterop()

	return nil
}

// Run starts the application
func (a *Application) Run() error {
	// Check if we need to authenticate
	if a.config.PlexToken == "" {
		a.showLoginScreen()
	} else {
		// Validate token and show main UI if valid
		go func() {
			if err := a.validatePlexToken(); err != nil {
				log.Printf("Token validation failed: %v", err)
				// Show login screen on main thread
				a.fyneApp.Driver().Run(func() {
					a.showLoginScreen()
				})
			} else {
				a.isAuthenticated = true
				a.fyneApp.Driver().Run(func() {
					a.showMainUI()
				})
			}
		}()
	}

	// Show loading screen initially
	a.showLoadingScreen("Starting Plex Desktop Client...")

	// Start the application
	a.mainWindow.ShowAndRun()

	// Clean up before exit
	a.cleanup()

	return nil
}

// validatePlexToken checks if the stored token is valid
func (a *Application) validatePlexToken() error {
	return a.plexClient.ValidateToken()
}

// showLoginScreen displays the login UI
func (a *Application) showLoginScreen() {
	loginScreen := a.uiController.NewLoginScreen(func(username, password string) {
		// Show loading dialog during login
		d := dialog.NewProgress("Logging in", "Authenticating with Plex...", a.mainWindow)
		d.Show()

		// Try to login in a goroutine
		go func() {
			token, err := a.plexClient.Login(username, password)
			// Update UI on main thread
			a.fyneApp.Driver().Run(func() {
				d.Hide()

				if err != nil {
					dialog.ShowError(err, a.mainWindow)
					return
				}

				// Store token in config
				a.config.PlexUsername = username
				a.config.PlexToken = token
				if err := a.config.Save(); err != nil {
					log.Printf("Failed to save config: %v", err)
				}

				a.isAuthenticated = true
				a.showMainUI()
			})
		}()
	})

	a.mainWindow.SetContent(loginScreen)
}

// showLoadingScreen displays a loading screen with the given message
func (a *Application) showLoadingScreen(message string) {
	a.mainWindow.SetContent(a.uiController.NewLoadingScreen(message))
}

// showMainUI displays the main application UI
func (a *Application) showMainUI() {
	// Initialize UI components
	sidebar := a.uiController.NewSidebar(
		func() { a.loadHomeScreen() },
		func() { a.loadLibrariesScreen() },
		func() { a.loadSearchScreen() },
		func() { a.loadSettingsScreen() },
	)

	// Create content container (initially empty)
	content := container.NewMax()

	// Create main layout (sidebar + content)
	mainLayout := container.NewBorder(nil, nil, sidebar, nil, content)

	// Set window content
	a.mainWindow.SetContent(mainLayout)

	// Initialize with home screen
	a.loadHomeScreen()

	// Start sync client
	if a.config.SyncEnabled {
		go a.syncClient.Start(a.config.PlexToken, func(session *sync.MediaSession) {
			// Handle remote sync events
			a.handleRemoteSync(session)
		})
	}
}

// loadHomeScreen shows the home screen with continue watching, etc
func (a *Application) loadHomeScreen() {
	a.showLoadingScreen("Loading home screen...")

	go func() {
		// Fetch data from Plex
		continueWatching, err1 := a.plexClient.GetContinueWatching()
		recentlyAdded, err2 := a.plexClient.GetRecentlyAdded()

		// Update UI on main thread
		a.fyneApp.Driver().Run(func() {
			if err1 != nil || err2 != nil {
				dialog.ShowError(fmt.Errorf("failed to load content"), a.mainWindow)
				return
			}

			homeScreen := a.uiController.NewHomeScreen(
				continueWatching,
				recentlyAdded,
				func(media *plex.MediaItem) {
					// Handle media selection
					a.playMedia(media)
				},
			)

			contentContainer := a.getContentContainer()
			if contentContainer != nil {
				contentContainer.Objects = []fyne.CanvasObject{homeScreen}
				contentContainer.Refresh()
			}
		})
	}()
}

// loadLibrariesScreen shows the libraries screen
func (a *Application) loadLibrariesScreen() {
	a.showLoadingScreen("Loading libraries...")

	go func() {
		// Fetch libraries from Plex
		libraries, err := a.plexClient.GetLibraries()

		// Update UI on main thread
		a.fyneApp.Driver().Run(func() {
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to load libraries: %v", err), a.mainWindow)
				return
			}

			librariesScreen := a.uiController.NewLibrariesScreen(
				libraries,
				func(library *plex.Library) {
					// Handle library selection
					a.loadLibraryContent(library)
				},
			)

			contentContainer := a.getContentContainer()
			if contentContainer != nil {
				contentContainer.Objects = []fyne.CanvasObject{librariesScreen}
				contentContainer.Refresh()
			}
		})
	}()
}

// loadLibraryContent shows the content of a specific library
func (a *Application) loadLibraryContent(library *plex.Library) {
	a.showLoadingScreen(fmt.Sprintf("Loading %s...", library.Title))

	go func() {
		// Fetch library content from Plex
		items, err := a.plexClient.GetLibraryItems(library.Key)

		// Update UI on main thread
		a.fyneApp.Driver().Run(func() {
			if err != nil {
				dialog.ShowError(fmt.Errorf("failed to load library content: %v", err), a.mainWindow)
				return
			}

			libraryScreen := a.uiController.NewLibraryContentScreen(
				library,
				items,
				func(media *plex.MediaItem) {
					// Handle media selection
					a.playMedia(media)
				},
			)

			contentContainer := a.getContentContainer()
			if contentContainer != nil {
				contentContainer.Objects = []fyne.CanvasObject{libraryScreen}
				contentContainer.Refresh()
			}
		})
	}()
}

// loadSearchScreen shows the search screen
func (a *Application) loadSearchScreen() {
	searchScreen := a.uiController.NewSearchScreen(func(query string) {
		// Handle search
		a.performSearch(query)
	})

	contentContainer := a.getContentContainer()
	if contentContainer != nil {
		contentContainer.Objects = []fyne.CanvasObject{searchScreen}
		contentContainer.Refresh()
	}
}

// performSearch searches for media with the given query
func (a *Application) performSearch(query string) {
	a.showLoadingScreen(fmt.Sprintf("Searching for %s...", query))

	go func() {
		// Search in Plex
		results, err := a.plexClient.Search(query)

		// Update UI on main thread
		a.fyneApp.Driver().Run(func() {
			if err != nil {
				dialog.ShowError(fmt.Errorf("search failed: %v", err), a.mainWindow)
				return
			}

			resultsScreen := a.uiController.NewSearchResultsScreen(
				query,
				results,
				func(media *plex.MediaItem) {
					// Handle media selection
					a.playMedia(media)
				},
			)

			contentContainer := a.getContentContainer()
			if contentContainer != nil {
				contentContainer.Objects = []fyne.CanvasObject{resultsScreen}
				contentContainer.Refresh()
			}
		})
	}()
}

// loadSettingsScreen shows the settings screen
func (a *Application) loadSettingsScreen() {
	settingsScreen := a.uiController.NewSettingsScreen(a.config, func() {
		// Handle settings save
		if err := a.config.Save(); err != nil {
			dialog.ShowError(fmt.Errorf("failed to save settings: %v", err), a.mainWindow)
			return
		}

		dialog.ShowInformation("Settings", "Settings saved successfully", a.mainWindow)

		// Apply new settings
		applyTheme(a.fyneApp, a.config)
		a.player.UpdateConfig(&player.Config{
			DefaultVolume:    a.config.DefaultVolume,
			HardwareAccel:    a.config.EnableHardwareAccel,
			SubtitleFontSize: a.config.SubtitleFontSize,
		})
		a.syncClient.UpdateConfig(&sync.Config{
			ServerURL:    a.config.SyncServerURL,
			SyncInterval: a.config.SyncInterval,
			Enabled:      a.config.SyncEnabled,
		})
	})

	contentContainer := a.getContentContainer()
	if contentContainer != nil {
		contentContainer.Objects = []fyne.CanvasObject{settingsScreen}
		contentContainer.Refresh()
	}
}

// playMedia plays the selected media item
func (a *Application) playMedia(media *plex.MediaItem) {
	a.showLoadingScreen(fmt.Sprintf("Loading %s...", media.Title))

	go func() {
		// Get stream URL and metadata
		streamURL, err := a.plexClient.GetStreamURL(media.Key)
		if err != nil {
			a.fyneApp.Driver().Run(func() {
				dialog.ShowError(fmt.Errorf("failed to get stream URL: %v", err), a.mainWindow)
			})
			return
		}

		// Get playback position
		position := time.Duration(0)
		if media.ViewOffset > 0 {
			position = time.Duration(media.ViewOffset) * time.Millisecond
		}

		// Update UI on main thread
		a.fyneApp.Driver().Run(func() {
			// Set current media
			a.currentMedia = media

			// Play media
			if err := a.player.Play(streamURL, position); err != nil {
				dialog.ShowError(fmt.Errorf("playback failed: %v", err), a.mainWindow)
				return
			}

			// Show player screen
			playerScreen := a.uiController.NewPlayerScreen(
				media,
				a.player,
				func() {
					// Handle back button
					a.stopPlayback()
					a.loadHomeScreen()
				},
			)

			contentContainer := a.getContentContainer()
			if contentContainer != nil {
				contentContainer.Objects = []fyne.CanvasObject{playerScreen}
				contentContainer.Refresh()
			}

			// Start position reporting for sync
			if a.config.SyncEnabled {
				a.startPositionReporting()
			}
		})
	}()
}

// startPositionReporting reports playback position for sync
func (a *Application) startPositionReporting() {
	if a.currentMedia == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(a.config.SyncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if a.player.IsStopped() {
					return // Stop reporting if playback stopped
				}

				position := a.player.GetPosition()
				state := a.player.GetState()

				// Report to sync server
				a.syncClient.ReportPosition(&sync.MediaSession{
					MediaKey: a.currentMedia.Key,
					Position: int(position.Milliseconds()),
					Duration: int(a.player.GetDuration().Milliseconds()),
					State:    state,
				})

				// Also update Plex server directly
				a.plexClient.UpdatePlaybackProgress(
					a.currentMedia.Key,
					int(position.Milliseconds()),
					state,
				)
			}
		}
	}()
}

// handleRemoteSync handles remote sync events
func (a *Application) handleRemoteSync(session *sync.MediaSession) {
	// Only handle events for currently playing media
	if a.currentMedia == nil || a.currentMedia.Key != session.MediaKey {
		return
	}

	// Update player position/state
	a.fyneApp.Driver().Run(func() {
		position := time.Duration(session.Position) * time.Millisecond

		switch session.State {
		case "playing":
			a.player.Seek(position)
			a.player.Resume()
		case "paused":
			a.player.Seek(position)
			a.player.Pause()
		case "stopped":
			a.stopPlayback()
			a.loadHomeScreen()
		}
	})
}

// stopPlayback stops the current playback
func (a *Application) stopPlayback() {
	if a.player != nil {
		a.player.Stop()
	}
	a.currentMedia = nil
}

// getContentContainer gets the content container from the main layout
func (a *Application) getContentContainer() *container.Max {
	// Get the main layout (border container)
	mainLayout, ok := a.mainWindow.Content().(*container.Border)
	if !ok {
		return nil
	}

	// Get the content container (max container)
	contentContainer, ok := mainLayout.Center.(*container.Max)
	if !ok {
		return nil
	}

	return contentContainer
}

// setupInterop sets up interoperability between components
func (a *Application) setupInterop() {
	// Connect player events to UI updates
	a.player.OnStateChanged(func(state string) {
		// Update UI based on player state
		a.uiController.UpdatePlayerState(state)

		// Report state to sync if needed
		if a.config.SyncEnabled && a.currentMedia != nil {
			position := a.player.GetPosition()
			a.syncClient.ReportPosition(&sync.MediaSession{
				MediaKey: a.currentMedia.Key,
				Position: int(position.Milliseconds()),
				Duration: int(a.player.GetDuration().Milliseconds()),
				State:    state,
			})
		}
	})
}

// cleanup performs cleanup before application exit
func (a *Application) cleanup() {
	// Stop player
	if a.player != nil {
		a.player.Stop()
	}

	// Stop sync client
	if a.syncClient != nil {
		a.syncClient.Stop()
	}

	// Save any unsaved config
	if a.config != nil {
		_ = a.config.Save()
	}
}

// applyTheme applies the theme from config to the application
func applyTheme(a fyne.App, cfg *config.Config) {
	// Set theme based on config
	switch cfg.UITheme {
	case "light":
		a.Settings().SetTheme(theme.LightTheme())
	case "dark":
		a.Settings().SetTheme(theme.DarkTheme())
	default: // "system"
		// Use system theme (Fyne default)
	}
}
