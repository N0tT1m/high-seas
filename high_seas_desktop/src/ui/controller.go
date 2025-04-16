// src/ui/controller.go

package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/N0tT1m/sync-stream/src/player"
	"github.com/N0tT1m/sync-stream/src/plex"
)

// Controller manages the UI components
type Controller struct {
	app         fyne.App
	mainWindow  fyne.Window
	player      *player.Player
	playerState string
}

// NewController creates a new UI controller
func NewController(app fyne.App, mainWindow fyne.Window, p *player.Player) *Controller {
	controller := &Controller{
		app:         app,
		mainWindow:  mainWindow,
		player:      p,
		playerState: "stopped",
	}

	return controller
}

// UpdatePlayerState updates the player state
func (c *Controller) UpdatePlayerState(state string) {
	c.playerState = state
}

// NewLoadingScreen creates a loading screen
func (c *Controller) NewLoadingScreen(message string) fyne.CanvasObject {
	// Create progress spinner
	spinner := widget.NewProgressBarInfinite()

	// Create message label
	label := widget.NewLabel(message)
	label.Alignment = fyne.TextAlignCenter

	// Create layout
	return container.NewCenter(
		container.NewVBox(
			spinner,
			label,
		),
	)
}

// NewLoginScreen creates a login screen
func (c *Controller) NewLoginScreen(onLogin func(username, password string)) fyne.CanvasObject {
	// Create form
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Your Plex Username/Email")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Your Plex Password")

	loginForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Username", Widget: usernameEntry},
			{Text: "Password", Widget: passwordEntry},
		},
		OnSubmit: func() {
			onLogin(usernameEntry.Text, passwordEntry.Text)
		},
		SubmitText: "Sign In",
	}

	// Create logo
	logo := canvas.NewText("Plex Desktop", color.NRGBA{R: 229, G: 160, B: 13, A: 255})
	logo.TextSize = 24
	logo.Alignment = fyne.TextAlignCenter

	// Create layout
	return container.NewCenter(
		container.NewVBox(
			logo,
			widget.NewLabel(""), // Spacer
			loginForm,
		),
	)
}

// NewSidebar creates the navigation sidebar
func (c *Controller) NewSidebar(
	onHomeClick func(),
	onLibrariesClick func(),
	onSearchClick func(),
	onSettingsClick func(),
) fyne.CanvasObject {
	// Create buttons
	homeBtn := widget.NewButtonWithIcon("Home", theme.HomeIcon(), onHomeClick)
	librariesBtn := widget.NewButtonWithIcon("Libraries", theme.FolderIcon(), onLibrariesClick)
	searchBtn := widget.NewButtonWithIcon("Search", theme.SearchIcon(), onSearchClick)
	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), onSettingsClick)

	// Create layout
	return container.NewVBox(
		homeBtn,
		librariesBtn,
		searchBtn,
		layout.NewSpacer(),
		settingsBtn,
	)
}

// NewLibrariesScreen creates the libraries screen
func (c *Controller) NewLibrariesScreen(
	libraries []*plex.Library,
	onLibrarySelect func(*plex.Library),
) fyne.CanvasObject {
	// Create title
	title := canvas.NewText("Libraries", color.White)
	title.TextSize = 24

	// Create library list
	libraryList := container.NewVBox()
	if len(libraries) == 0 {
		libraryList.Add(widget.NewLabel("No libraries found"))
	} else {
		for _, lib := range libraries {
			libraryCard := NewLibraryCard(lib, onLibrarySelect)
			libraryList.Add(libraryCard)
		}
	}

	// Create layout
	return container.NewVBox(
		title,
		widget.NewSeparator(),
		container.NewVScroll(libraryList),
	)
}

// NewLibraryContentScreen creates a screen showing library content
func (c *Controller) NewLibraryContentScreen(
	library *plex.Library,
	items []*plex.MediaItem,
	onMediaSelect func(*plex.MediaItem),
) fyne.CanvasObject {
	// Create title
	title := canvas.NewText(library.Title, color.White)
	title.TextSize = 24

	// Create grid layout for media items
	grid := container.NewGridWrap(fyne.NewSize(180, 270))

	if len(items) == 0 {
		grid.Add(widget.NewLabel("No items found"))
	} else {
		for _, item := range items {
			mediaCard := NewMediaCard(item, onMediaSelect)
			grid.Add(mediaCard)
		}
	}

	// Create layout
	return container.NewVBox(
		title,
		widget.NewSeparator(),
		container.NewVScroll(grid),
	)
}

// NewHomeScreen creates the home screen
func (c *Controller) NewHomeScreen(
	continueWatching []*plex.MediaItem,
	recentlyAdded []*plex.MediaItem,
	onMediaSelect func(*plex.MediaItem),
) fyne.CanvasObject {
	// Create title
	title := canvas.NewText("Home", color.White)
	title.TextSize = 24

	// Create continue watching section
	cwTitle := canvas.NewText("Continue Watching", color.White)
	cwTitle.TextSize = 18

	cwList := container.NewHBox()
	if len(continueWatching) == 0 {
		cwList.Add(widget.NewLabel("No items to continue watching"))
	} else {
		for _, item := range continueWatching {
			mediaCard := NewMediaCard(item, onMediaSelect)
			cwList.Add(mediaCard)
		}
	}

	cwSection := container.NewVBox(
		cwTitle,
		container.NewHScroll(cwList),
	)

	// Create recently added section
	raTitle := canvas.NewText("Recently Added", color.White)
	raTitle.TextSize = 18

	raList := container.NewHBox()
	if len(recentlyAdded) == 0 {
		raList.Add(widget.NewLabel("No recently added items"))
	} else {
		for _, item := range recentlyAdded {
			mediaCard := NewMediaCard(item, onMediaSelect)
			raList.Add(mediaCard)
		}
	}

	raSection := container.NewVBox(
		raTitle,
		container.NewHScroll(raList),
	)

	// Create layout
	return container.NewVBox(
		title,
		widget.NewSeparator(),
		cwSection,
		widget.NewSeparator(),
		raSection,
	)
}
