package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	PlexAPIURL = "http://192.168.1.78:5000"
)

type PlexResponse struct {
	Status      string   `json:"status"`
	Error       string   `json:"error"`
	QueueLength int      `json:"queue_length"`
	Items       []string `json:"items"`
	Movies      int      `json:"movies"` // Add this field
	Shows       int      `json:"shows"`  // Add this field
}

var httpClient = &http.Client{
	Timeout: 60 * time.Second, // Increased timeout
	Transport: &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    true,
		ResponseHeaderTimeout: 30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
	},
}

type listMediaResponse struct {
	Items []struct {
		Title        string  `json:"title"`
		Year         int     `json:"year"`
		Summary      string  `json:"summary"`
		Rating       float64 `json:"rating"`
		Duration     int     `json:"duration"`
		EpisodeCount int     `json:"episode_count,omitempty"`
		SeasonCount  int     `json:"season_count,omitempty"`
	} `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
}

// Define result struct globally
type searchResult struct {
	Results []struct {
		Title   string `json:"title"`
		Year    int    `json:"year"`
		Summary string `json:"summary"`
	} `json:"results"`
	Count int `json:"count"`
}

var (
	commands        []*discordgo.ApplicationCommand
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}
)

func init() {
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "movies",
			Description: "Show unwatched movie count",
		},
		{
			Name:        "shows",
			Description: "Show unwatched show count",
		},
		{
			Name:        "smart-search",
			Description: "Search for media by name",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "term",
					Description: "Search term",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Media type",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Movies", Value: "movies"},
						{Name: "Shows", Value: "shows"},
					},
				},
			},
		},
		{
			Name:        "queue",
			Description: "Queue specific media",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Media type (movie/show)",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Movie", Value: "movie"},
						{Name: "Show", Value: "show"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "titles",
					Description: "Comma-separated titles",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "get_all",
					Description: "Queue all matching items instead of just the first match",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "use_existing_queue",
					Description: "Add to existing queue if available",
					Required:    false,
				},
			},
		},
		{
			Name:        "add",
			Description: "Add to existing queue",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Media type (movie/show)",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Movie", Value: "movie"},
						{Name: "Show", Value: "show"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "titles",
					Description: "Comma-separated titles",
					Required:    true,
				},
			},
		},
		{
			Name:        "list",
			Description: "List media from Plex",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Media type",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Movies", Value: "movies"},
						{Name: "Shows", Value: "shows"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "Page number",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "search",
					Description: "Search term",
					Required:    false,
				},
			},
		},
		{
			Name:        "player",
			Description: "Control media playback",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "action",
					Description: "Player action",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Play", Value: "play"},
						{Name: "Pause", Value: "pause"},
						{Name: "Stop", Value: "stop"},
						{Name: "Next", Value: "skipNext"},
						{Name: "Previous", Value: "skipPrevious"},
						{Name: "Mute", Value: "mute"},
						{Name: "Unmute", Value: "unmute"},
					},
				},
			},
		},
		{
			Name:        "volume",
			Description: "Set player volume",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "level",
					Description: "Volume level (0-100)",
					Required:    true,
					MinValue:    &[]float64{0}[0],
					MaxValue:    100,
				},
			},
		},
		{
			Name:        "seek",
			Description: "Seek to position",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "seconds",
					Description: "Position in seconds",
					Required:    true,
				},
			},
		},
		{
			Name:        "status",
			Description: "Show player status",
		},
		{
			Name:        "clients",
			Description: "List available Plex clients",
		},
		{
			Name:        "playlists",
			Description: "List Plex playlists",
		},
		{
			Name:        "queue-status",
			Description: "Show current queue status",
		},
		{
			Name:        "clear-queue",
			Description: "Clear the current queue",
		},
		{
			Name:        "random_movie",
			Description: "Queue random movies",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Number of movies to queue",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "use_existing_queue",
					Description: "Add to existing queue if available",
					Required:    false,
				},
			},
		},
		{
			Name:        "random_show",
			Description: "Queue random shows",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Number of shows to queue",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "use_existing_queue",
					Description: "Add to existing queue if available",
					Required:    false,
				},
			},
		},
	}

	// Update the queue-status command handler
	commandHandlers["queue-status"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Println("Starting queue-status command")
		resp, err := http.Get(PlexAPIURL + "/get-current-queue")
		if err != nil {
			log.Printf("Error getting queue status: %v", err)
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			respondError(s, i, err)
			return
		}

		var result struct {
			Status      string `json:"status"`
			Message     string `json:"message"`
			QueueLength int    `json:"queue_length"`
			CurrentItem string `json:"current_item"`
			Items       []struct {
				Title    string `json:"title"`
				Type     string `json:"type"`
				Duration int    `json:"duration"`
				Selected bool   `json:"selected"`
			} `json:"items"`
			Error string `json:"error"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("Error parsing response: %v", err)
			respondError(s, i, err)
			return
		}

		// If there's a message but no error, it's an info message (like "No active queue")
		if result.Message != "" && result.Error == "" {
			embed := &discordgo.MessageEmbed{
				Title:       "Queue Status",
				Description: result.Message,
				Color:       0x0099FF, // Blue for info
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
			return
		}

		if result.Error != "" {
			log.Printf("Error in response: %s", result.Error)
			respondError(s, i, fmt.Errorf(result.Error))
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Queue Status",
			Description: fmt.Sprintf("Total items in queue: %d", result.QueueLength),
			Fields:      make([]*discordgo.MessageEmbedField, 0),
			Color:       0x00FF00, // Green for success
		}

		if result.CurrentItem != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Now Playing",
				Value:  result.CurrentItem,
				Inline: false,
			})
		}

		for idx, item := range result.Items {
			duration := fmt.Sprintf("%d:%02d", item.Duration/60000, (item.Duration/1000)%60)
			status := ""
			if item.Selected {
				status = " (Current)"
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%d. %s%s", idx+1, item.Title, status),
				Value:  fmt.Sprintf("Type: %s\nDuration: %s", item.Type, duration),
				Inline: true,
			})
		}

		log.Println("Sending queue status response")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	// Update the clear-queue command handler
	commandHandlers["clear-queue"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Println("Starting clear-queue command")
		resp, err := http.Post(PlexAPIURL+"/clear-queue", "application/json", nil)
		if err != nil {
			log.Printf("Error clearing queue: %v", err)
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			respondError(s, i, err)
			return
		}

		var result struct {
			Message string `json:"message"`
			Status  string `json:"status"`
			Error   string `json:"error"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("Error parsing response: %v", err)
			respondError(s, i, err)
			return
		}

		if result.Error != "" {
			log.Printf("Error in response: %s", result.Error)
			respondError(s, i, fmt.Errorf(result.Error))
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Clear Queue",
			Description: result.Message,
			Color:       0x00FF00, // Green for success
		}

		log.Println("Sending clear queue response")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	commandHandlers["random_movie"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleRandomMedia(s, i, "movie")
	}

	commandHandlers["random_show"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleRandomMedia(s, i, "show")
	}

	commandHandlers["list"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleList(s, i)
	}

	// Update the queue command handler
	// Update the queue command handler
	commandHandlers["queue"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Println("Starting queue command handler")

		// Immediately acknowledge the command
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Processing your request...",
			},
		})
		if err != nil {
			log.Printf("Error acknowledging interaction: %v", err)
			return
		}

		// Process command parameters
		mediaType := i.ApplicationCommandData().Options[0].StringValue()
		titles := strings.Split(i.ApplicationCommandData().Options[1].StringValue(), ",")
		for i := range titles {
			titles[i] = strings.TrimSpace(titles[i])
		}

		getAll := false
		useExistingQueue := false
		if len(i.ApplicationCommandData().Options) > 2 {
			getAll = i.ApplicationCommandData().Options[2].BoolValue()
		}
		if len(i.ApplicationCommandData().Options) > 3 {
			useExistingQueue = i.ApplicationCommandData().Options[3].BoolValue()
		}

		log.Printf("Processed parameters - Type: %s, Titles: %v, GetAll: %v, UseExistingQueue: %v",
			mediaType, titles, getAll, useExistingQueue)

		// Prepare request
		data := map[string]interface{}{
			"type":               mediaType,
			"items":              titles,
			"get_all":            getAll,
			"use_existing_queue": useExistingQueue,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshaling data: %v", err)
			sendFollowUpError(s, i, fmt.Errorf("failed to prepare request: %v", err))
			return
		}

		log.Printf("Sending request to %s with data: %s", PlexAPIURL+"/queue-specific", string(jsonData))

		// Make HTTP request with timeout
		req, err := http.NewRequest("POST", PlexAPIURL+"/queue-specific", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			sendFollowUpError(s, i, fmt.Errorf("failed to create request: %v", err))
			return
		}
		req.Header.Set("Content-Type", "application/json")

		// Make the request
		resp, err := httpClient.Do(req)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				log.Printf("Request timed out: %v", err)
				sendFollowUpError(s, i, fmt.Errorf("request timed out, please try again"))
			} else {
				log.Printf("Request failed: %v", err)
				sendFollowUpError(s, i, fmt.Errorf("request failed: %v", err))
			}
			return
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			sendFollowUpError(s, i, fmt.Errorf("failed to read response: %v", err))
			return
		}

		log.Printf("Raw response from server: %s", string(body))

		var result PlexResponse
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("Error unmarshaling response: %v", err)
			sendFollowUpError(s, i, fmt.Errorf("failed to parse response: %v", err))
			return
		}

		log.Printf("Parsed response: %+v", result)

		if result.Error != "" {
			log.Printf("Error in response: %s", result.Error)
			sendFollowUpError(s, i, fmt.Errorf(result.Error))
			return
		}

		// Create embed
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Queue %s", strings.Title(mediaType)),
			Description: fmt.Sprintf("Added %d items to queue:", result.QueueLength),
			Fields:      make([]*discordgo.MessageEmbedField, 0),
			Color:       0x00FF00, // Green for success
		}

		for idx, title := range result.Items {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%d.", idx+1),
				Value:  title,
				Inline: true,
			})
		}

		log.Printf("Created embed: %+v", embed)

		// Send follow-up message
		_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
		if err != nil {
			log.Printf("Error sending follow-up message: %v", err)
			sendFollowUpError(s, i, fmt.Errorf("failed to send response: %v", err))
			return
		}

		log.Println("Queue command completed successfully")
	}

	// Update the movies command
	commandHandlers["movies"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/movies")
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		var plexResp PlexResponse
		if err := json.NewDecoder(resp.Body).Decode(&plexResp); err != nil {
			respondError(s, i, err)
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Unwatched Movies",
			Description: fmt.Sprintf("Found %d unwatched movies", plexResp.Movies),
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	// Update the shows command
	commandHandlers["shows"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/shows")
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		var plexResp PlexResponse
		if err := json.NewDecoder(resp.Body).Decode(&plexResp); err != nil {
			respondError(s, i, err)
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Unwatched Shows",
			Description: fmt.Sprintf("Found %d unwatched shows", plexResp.Shows),
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	// Update the add command to use the standardized format
	commandHandlers["add"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		mediaType := i.ApplicationCommandData().Options[0].StringValue()
		titles := strings.Split(i.ApplicationCommandData().Options[1].StringValue(), ",")
		for i := range titles {
			titles[i] = strings.TrimSpace(titles[i])
		}

		data := map[string]interface{}{
			"type":   mediaType,
			"titles": titles,
		}
		makePostRequest(s, i, "/add-to-queue", data)
	}

	// Fix for player command - properly declare and use jsonData
	commandHandlers["player"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		action := i.ApplicationCommandData().Options[0].StringValue()
		data := map[string]string{"action": action}

		jsonData, err := json.Marshal(data)
		if err != nil {
			respondError(s, i, err)
			return
		}

		resp, err := http.Post(PlexAPIURL+"/player-controls", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		embed := &discordgo.MessageEmbed{
			Title:       "Player Control",
			Description: fmt.Sprintf("Successfully executed action: %s", action),
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	// Fix for volume command - properly declare and use jsonData
	commandHandlers["volume"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		volume := i.ApplicationCommandData().Options[0].IntValue()
		volumeData := map[string]interface{}{
			"action": "setVolume",
			"volume": volume,
		}

		jsonData, err := json.Marshal(volumeData)
		if err != nil {
			respondError(s, i, err)
			return
		}

		resp, err := http.Post(PlexAPIURL+"/player-controls", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		embed := &discordgo.MessageEmbed{
			Title:       "Volume Control",
			Description: fmt.Sprintf("Set volume to: %d%%", volume),
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	// Fix for seek command - properly declare and use jsonData
	commandHandlers["seek"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		seconds := i.ApplicationCommandData().Options[0].IntValue()
		seekData := map[string]interface{}{
			"action": "seekTo",
			"time":   seconds * 1000, // Convert to milliseconds
		}

		jsonData, err := json.Marshal(seekData)
		if err != nil {
			respondError(s, i, err)
			return
		}

		resp, err := http.Post(PlexAPIURL+"/player-controls", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		embed := &discordgo.MessageEmbed{
			Title:       "Seek Command",
			Description: fmt.Sprintf("Seeked to position: %d:%02d", seconds/60, seconds%60),
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	// Update status command
	commandHandlers["status"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/player-status")
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		var status struct {
			State    string `json:"state"`
			Time     int    `json:"time"`
			Duration int    `json:"duration"`
			Volume   int    `json:"volume"`
			Muted    bool   `json:"muted"`
			Current  *struct {
				Title string `json:"title"`
				Type  string `json:"type"`
			} `json:"current_media"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			respondError(s, i, err)
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:  "Player Status",
			Fields: make([]*discordgo.MessageEmbedField, 0),
		}

		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{
				Name:   "State",
				Value:  strings.Title(status.State),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Volume",
				Value:  fmt.Sprintf("%d%% %s", status.Volume, map[bool]string{true: "(Muted)", false: ""}[status.Muted]),
				Inline: true,
			},
		)

		if status.Current != nil {
			embed.Fields = append(embed.Fields,
				&discordgo.MessageEmbedField{
					Name:   "Now Playing",
					Value:  status.Current.Title,
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "Progress",
					Value:  fmt.Sprintf("%d:%02d / %d:%02d", status.Time/60000, (status.Time/1000)%60, status.Duration/60000, (status.Duration/1000)%60),
					Inline: true,
				},
			)
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	// Update clients command
	commandHandlers["clients"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/get-clients")
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		var result struct {
			Clients []struct {
				Name     string `json:"name"`
				Device   string `json:"device"`
				Platform string `json:"platform"`
				State    string `json:"state"`
			} `json:"clients"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			respondError(s, i, err)
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Plex Clients",
			Description: fmt.Sprintf("Found %d client(s)", len(result.Clients)),
			Fields:      make([]*discordgo.MessageEmbedField, 0),
		}

		for _, client := range result.Clients {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name: client.Name,
				Value: fmt.Sprintf("Device: %s\nPlatform: %s\nState: %s",
					client.Device, client.Platform, client.State),
				Inline: true,
			})
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	// Update playlists command
	commandHandlers["playlists"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/get-playlists")
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		var result struct {
			Playlists []struct {
				Title    string `json:"title"`
				Items    int    `json:"items"`
				Duration int    `json:"duration"`
				Type     string `json:"type"`
			} `json:"playlists"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			respondError(s, i, err)
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Plex Playlists",
			Description: fmt.Sprintf("Found %d playlist(s)", len(result.Playlists)),
			Fields:      make([]*discordgo.MessageEmbedField, 0),
		}

		for _, playlist := range result.Playlists {
			hours := playlist.Duration / 3600000
			minutes := (playlist.Duration % 3600000) / 60000

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name: playlist.Title,
				Value: fmt.Sprintf("Type: %s\nItems: %d\nDuration: %dh %dm",
					playlist.Type, playlist.Items, hours, minutes),
				Inline: true,
			})
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	// Update smart-search command
	commandHandlers["smart-search"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		term := i.ApplicationCommandData().Options[0].StringValue()
		mediaType := i.ApplicationCommandData().Options[1].StringValue()

		searchURL := fmt.Sprintf("%s/smart-search?term=%s&type=%s",
			PlexAPIURL, url.QueryEscape(term), mediaType)

		resp, err := http.Get(searchURL)
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		var result searchResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			respondError(s, i, err)
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Search Results: %s", term),
			Description: fmt.Sprintf("Found %d matches", result.Count),
			Fields:      make([]*discordgo.MessageEmbedField, 0),
		}

		for _, item := range result.Results {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%s (%d)", item.Title, item.Year),
				Value:  item.Summary,
				Inline: false,
			})
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}
}

func makePostRequest(s *discordgo.Session, i *discordgo.InteractionCreate, endpoint string, data interface{}) {
	log.Printf("Making POST request to %s with data: %+v\n", endpoint, data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		respondError(s, i, err)
		return
	}

	resp, err := http.Post(PlexAPIURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error making request: %v", err)
		respondError(s, i, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		respondError(s, i, err)
		return
	}

	log.Printf("Raw response: %s", string(body))

	var result PlexResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		respondError(s, i, err)
		return
	}

	if result.Error != "" {
		log.Printf("Error in response: %s", result.Error)
		respondError(s, i, fmt.Errorf(result.Error))
		return
	}

	// Create standardized embed response
	embed := &discordgo.MessageEmbed{
		Title:       "Plex Command Result",
		Description: fmt.Sprintf("Operation completed successfully"),
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	if result.QueueLength > 0 {
		embed.Description = fmt.Sprintf("Added %d items to queue", result.QueueLength)
		for idx, item := range result.Items {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%d.", idx+1),
				Value:  item,
				Inline: true,
			})
		}
	}

	log.Println("Sending response to Discord")
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	if err != nil {
		log.Printf("Error sending Discord response: %v", err)
	}
}

// Update error response to use embeds
func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	log.Printf("Sending error response: %v", err)

	embed := &discordgo.MessageEmbed{
		Title:       "Error",
		Description: err.Error(),
		Color:       0xFF0000, // Red color for errors
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// Separate handler function
func handleSmartSearch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	term := i.ApplicationCommandData().Options[0].StringValue()
	mediaType := i.ApplicationCommandData().Options[1].StringValue()

	url := fmt.Sprintf("%s/smart-search?term=%s&type=%s",
		PlexAPIURL, url.QueryEscape(term), mediaType)

	resp, err := http.Get(url)
	if err != nil {
		respondError(s, i, err)
		return
	}
	defer resp.Body.Close()

	var result searchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		respondError(s, i, err)
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("Search Results for '%s' (%d found)", term, result.Count),
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, item := range result.Results {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%s (%d)", item.Title, item.Year),
			Value: item.Summary,
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// Main func
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken := os.Getenv("TOKEN")

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is ready")
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer dg.Close()

	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Error creating command %v: %v", cmd.Name, err)
		}
	}

	log.Println("Bot is running. Press CTRL-C to exit.")
	select {}
}

func handleList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("Starting handleList function")

	options := i.ApplicationCommandData().Options
	mediaType := options[0].StringValue()
	page := 1
	if len(options) > 1 {
		page = int(options[1].IntValue())
	}
	search := ""
	if len(options) > 2 {
		search = options[2].StringValue()
	}

	log.Printf("List parameters - type: %s, page: %d, search: %s", mediaType, page, search)

	baseURL := fmt.Sprintf("%s/list-media?type=%s&page=%d", PlexAPIURL, mediaType, page)
	if search != "" {
		baseURL += "&search=" + url.QueryEscape(search)
	}

	log.Printf("Making request to: %s", baseURL)
	resp, err := http.Get(baseURL)
	if err != nil {
		log.Printf("Error making request: %v", err)
		respondError(s, i, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		respondError(s, i, err)
		return
	}

	var result listMediaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		respondError(s, i, err)
		return
	}

	log.Printf("Got %d items from response", len(result.Items))

	searchInfo := ""
	if search != "" {
		searchInfo = fmt.Sprintf(" - Search: %s", search)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Library%s", strings.Title(mediaType), searchInfo),
		Description: fmt.Sprintf("Page %d of %d (Total items: %d)", result.Page, result.TotalPages, result.Total),
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	for _, item := range result.Items {
		// Format duration in hours and minutes
		hours := item.Duration / 3600000
		minutes := (item.Duration % 3600000) / 60000

		description := fmt.Sprintf("Year: %d\nRating: %.1f\nDuration: ", item.Year, item.Rating)
		if hours > 0 {
			description += fmt.Sprintf("%dh ", hours)
		}
		description += fmt.Sprintf("%dm", minutes)

		if mediaType == "shows" {
			description += fmt.Sprintf("\nSeasons: %d\nEpisodes: %d",
				item.SeasonCount, item.EpisodeCount)
		}

		if item.Summary != "" {
			// Truncate summary if it's too long
			summary := item.Summary
			if len(summary) > 100 {
				summary = summary[:97] + "..."
			}
			description += fmt.Sprintf("\n\n%s", summary)
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   item.Title,
			Value:  description,
			Inline: true,
		})
	}

	// Add navigation footer if multiple pages
	if result.TotalPages > 1 {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Use /list type:%s page:[1-%d] to navigate", mediaType, result.TotalPages),
		}
	}

	log.Println("Sending response to Discord")
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	if err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

func handleRandomMedia(s *discordgo.Session, i *discordgo.InteractionCreate, mediaType string) {
	log.Println("Starting handleRandomMedia function")

	// Immediately acknowledge the interaction
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Processing your request...",
		},
	})
	if err != nil {
		log.Printf("Error acknowledging interaction: %v", err)
		return
	}

	options := i.ApplicationCommandData().Options
	count := options[0].IntValue()
	useExistingQueue := false
	if len(options) > 1 {
		useExistingQueue = options[1].BoolValue()
	}

	log.Printf("Parameters - count: %d, mediaType: %s, useExistingQueue: %v", count, mediaType, useExistingQueue)

	data := map[string]interface{}{
		"number":             count,
		"type":               mediaType,
		"use_existing_queue": useExistingQueue,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		sendFollowUpError(s, i, err)
		return
	}

	log.Printf("Making POST request to %s/get-random-media with data: %s", PlexAPIURL, string(jsonData))

	req, err := http.NewRequest("POST", PlexAPIURL+"/get-random-media", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		sendFollowUpError(s, i, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		sendFollowUpError(s, i, fmt.Errorf("failed to make request: %v", err))
		return
	}
	defer resp.Body.Close()

	log.Printf("Received response with status code: %d", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		sendFollowUpError(s, i, fmt.Errorf("failed to read response body: %v", err))
		return
	}

	log.Printf("Raw response body: %s", string(body))

	var result PlexResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		sendFollowUpError(s, i, fmt.Errorf("failed to parse response: %v\nraw response: %s", err, string(body)))
		return
	}

	log.Printf("Parsed response: %+v", result)

	if result.Error != "" {
		log.Printf("Error in response: %s", result.Error)
		sendFollowUpError(s, i, fmt.Errorf(result.Error))
		return
	}

	log.Println("Creating embed message")
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Random %s Queue", strings.Title(mediaType)),
		Description: fmt.Sprintf("Added %d items to queue", result.QueueLength),
		Fields:      make([]*discordgo.MessageEmbedField, 0),
		Color:       0x00FF00, // Add green color for success
	}

	for idx, title := range result.Items {
		log.Printf("Adding item to embed: %s", title)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d.", idx+1),
			Value:  title,
			Inline: true,
		})
	}

	log.Println("Sending follow-up message to Discord")
	// Use FollowupMessageCreate instead of InteractionRespond
	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	})

	if err != nil {
		log.Printf("Error sending follow-up message: %v", err)
		sendFollowUpError(s, i, fmt.Errorf("failed to send response: %v", err))
	} else {
		log.Println("Successfully sent Discord response")
	}
}

func sendFollowUpError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	errorEmbed := &discordgo.MessageEmbed{
		Title:       "Error",
		Description: err.Error(),
		Color:       0xFF0000,
	}

	// Use FollowupMessageCreate instead of InteractionResponseEdit
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{errorEmbed},
	})
	if err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}
