// src/ui/widgets.go

package ui

import (
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/N0tT1m/sync-stream/src/plex"
)

// MediaCard creates a card for displaying a media item
func NewMediaCard(item *plex.MediaItem, onSelect func(*plex.MediaItem)) fyne.CanvasObject {
	// Create container with fixed size
	card := container.NewVBox()
	card.Resize(fyne.NewSize(160, 240))

	// Create poster image
	var poster *canvas.Image
	if item.Thumb != "" {
		// In a real implementation, we'd load the image from the Plex server
		// For this example, we'll use a placeholder
		poster = canvas.NewImageFromResource(fyne.NewStaticResource("poster", []byte{}))
		poster.FillMode = canvas.ImageFillContain
		poster.SetMinSize(fyne.NewSize(160, 230))

		// In a real implementation, we'd load the image asynchronously
		// This is just a placeholder for how that might work
		go func() {
			// Placeholder for image loading - in real code you'd fetch from Plex
			// imageURL := fmt.Sprintf("%s%s?X-Plex-Token=%s", serverURL, item.Thumb, token)
			// resp, err := http.Get(imageURL)
			// if err == nil && resp.StatusCode == http.StatusOK {
			//     // Load image data
			//     // Update poster.Resource
			//     poster.Refresh()
			// }
			// Just simulate loading delay for example
			time.Sleep(300 * time.Millisecond)
		}()
	} else {
		// Use placeholder for items without thumbnails
		rect := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 40, A: 255})
		rect.SetMinSize(fyne.NewSize(160, 230))
		poster = &canvas.Image{}
		poster.SetMinSize(fyne.NewSize(160, 230))
	}

	// Create title label
	title := widget.NewLabelWithStyle(
		truncateString(item.Title, 16),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Create subtitle label based on media type
	var subtitle *widget.Label
	if item.Type == "movie" {
		if item.Year > 0 {
			subtitle = widget.NewLabelWithStyle(
				truncateString(string(item.Year), 20),
				fyne.TextAlignCenter,
				fyne.TextStyle{},
			)
		} else {
			subtitle = widget.NewLabel("")
		}
	} else if item.Type == "episode" {
		subtitle = widget.NewLabelWithStyle(
			truncateString(item.GrandparentTitle, 16),
			fyne.TextAlignCenter,
			fyne.TextStyle{},
		)
	} else if item.Type == "album" {
		subtitle = widget.NewLabelWithStyle(
			truncateString(item.ParentTitle, 16),
			fyne.TextAlignCenter,
			fyne.TextStyle{},
		)
	} else {
		subtitle = widget.NewLabel("")
	}

	// Create progress overlay for in-progress items
	var progressOverlay fyne.CanvasObject
	if item.ViewOffset > 0 && item.Duration > 0 {
		progress := float32(item.ViewOffset) / float32(item.Duration)
		progressBar := widget.NewProgressBar()
		progressBar.Value = float64(progress)
		progressOverlay = progressBar
	} else {
		progressOverlay = widget.NewLabel("")
	}

	// Combine components
	content := container.NewVBox(
		container.NewMax(
			poster,
			container.NewVBox(
				layout.NewSpacer(),
				progressOverlay,
			),
		),
		title,
		subtitle,
	)

	// Wrap in a button to handle clicks
	button := widget.NewButton("", func() {
		onSelect(item)
	})
	button.Importance = widget.LowImportance

	// Create a BorderContainer with the content and transparent button on top
	return container.NewMax(
		content,
		button,
	)
}

// LibraryCard creates a card for displaying a library
func NewLibraryCard(library *plex.Library, onSelect func(*plex.Library)) fyne.CanvasObject {
	// Create icon based on library type
	var icon fyne.Resource
	switch library.Type {
	case "movie":
		icon = theme.FileVideoIcon()
	case "show":
		icon = theme.FileVideoIcon()
	case "artist":
		icon = theme.FileAudioIcon()
	case "photo":
		icon = theme.FileImageIcon()
	default:
		icon = theme.FolderIcon()
	}

	// Create button with icon and title
	button := widget.NewButtonWithIcon(library.Title, icon, func() {
		onSelect(library)
	})

	return button
}

// Helper functions

// truncateString truncates a string to the given length and adds ellipsis
func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
