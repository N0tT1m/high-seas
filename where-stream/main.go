package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	PlexAPIURL = "http://localhost:5000"
)

type plexResponse struct {
	Status      string   `json:"status"`
	Error       string   `json:"error"`
	QueueLength int      `json:"queue_length"`
	Items       []string `json:"items"`
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
			Name:        "random_movie",
			Description: "Queue random movies",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Number of movies to queue",
					Required:    true,
				},
			},
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
			Name:        "random_show",
			Description: "Queue random shows",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Number of shows to queue",
					Required:    true,
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
	}

	commandHandlers["movies"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/movies")
		if err != nil {
			respondError(s, i, err)
			return
		}

		var plexResp PlexResponse
		if err := json.NewDecoder(resp.Body).Decode(&plexResp); err != nil {
			respondError(s, i, err)
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Found %d unwatched movies", plexResp.Movies),
			},
		})
	}

	commandHandlers["shows"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/shows")
		if err != nil {
			respondError(s, i, err)
			return
		}

		var plexResp PlexResponse
		if err := json.NewDecoder(resp.Body).Decode(&plexResp); err != nil {
			respondError(s, i, err)
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Found %d unwatched shows", plexResp.Shows),
			},
		})
	}

	commandHandlers["random_movie"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleRandomMedia(s, i, "movie")
	}

	commandHandlers["random_show"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleRandomMedia(s, i, "show")
	}

	commandHandlers["queue"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		mediaType := i.ApplicationCommandData().Options[0].StringValue()
		titles := strings.Split(i.ApplicationCommandData().Options[1].StringValue(), ",")
		for i := range titles {
			titles[i] = strings.TrimSpace(titles[i])
		}

		data := map[string]interface{}{
			"type":  mediaType,
			"items": titles,
		}

		// Log the request data
		fmt.Printf("Sending request to %s/queue-specific with data: %+v\n", PlexAPIURL, data)

		jsonData, err := json.Marshal(data)
		if err != nil {
			respondError(s, i, err)
			return
		}

		resp, err := http.Post(PlexAPIURL+"/queue-specific", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			respondError(s, i, err)
			return
		}
		defer resp.Body.Close()

		var result struct {
			QueueLength int      `json:"queue_length"`
			Items       []string `json:"items"`
			Error       string   `json:"error"`
		}

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response from server: %s\n", string(body))

		if err := json.Unmarshal(body, &result); err != nil {
			respondError(s, i, fmt.Errorf("failed to parse response: %v", err))
			return
		}

		if result.Error != "" {
			respondError(s, i, fmt.Errorf(result.Error))
			return
		}

		message := fmt.Sprintf("Queued %d items:\n%s", result.QueueLength, strings.Join(result.Items, "\n"))
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
			},
		})
	}

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

	"list": handleList,

	commandHandlers["smart-search"] = handleSmartSearch

	// Add handlers for new commands
	commandHandlers["player"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		action := i.ApplicationCommandData().Options[0].StringValue()
		data := map[string]string{"action": action}
		makePostRequest(s, i, "/player-controls", data)
	}

	commandHandlers["volume"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		volume := i.ApplicationCommandData().Options[0].IntValue()
		data := map[string]interface{}{
			"action": "setVolume",
			"volume": volume,
		}
		makePostRequest(s, i, "/player-controls", data)
	}

	commandHandlers["seek"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		seconds := i.ApplicationCommandData().Options[0].IntValue()
		data := map[string]interface{}{
			"action": "seekTo",
			"time":   seconds * 1000, // Convert to milliseconds
		}
		makePostRequest(s, i, "/player-controls", data)
	}

	commandHandlers["status"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/player-status")
		if err != nil {
			respondError(s, i, err)
			return
		}

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
			Title: "Player Status",
			Fields: []*discordgo.MessageEmbedField{
				{Name: "State", Value: status.State},
				{Name: "Volume", Value: fmt.Sprintf("%d%% %s", status.Volume, map[bool]string{true: "(Muted)", false: ""}[status.Muted])},
			},
		}

		if status.Current != nil {
			embed.Fields = append(embed.Fields,
				&discordgo.MessageEmbedField{Name: "Playing", Value: status.Current.Title},
				&discordgo.MessageEmbedField{Name: "Progress", Value: fmt.Sprintf("%d:%02d / %d:%02d",
					status.Time/60000, (status.Time/1000)%60,
					status.Duration/60000, (status.Duration/1000)%60)},
			)
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}

	commandHandlers["clients"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/get-clients")
		if err != nil {
			respondError(s, i, err)
			return
		}

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
			Title:  "Plex Clients",
			Fields: []*discordgo.MessageEmbedField{},
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

	commandHandlers["playlists"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		resp, err := http.Get(PlexAPIURL + "/get-playlists")
		if err != nil {
			respondError(s, i, err)
			return
		}

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
			Title:  "Plex Playlists",
			Fields: []*discordgo.MessageEmbedField{},
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

func makePostRequest(s *discordgo.Session, i *discordgo.InteractionCreate, endpoint string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		respondError(s, i, err)
		return
	}

	resp, err := http.Post(PlexAPIURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		respondError(s, i, err)
		return
	}

	var plexResp PlexResponse
	if err := json.NewDecoder(resp.Body).Decode(&plexResp); err != nil {
		respondError(s, i, err)
		return
	}

	message := fmt.Sprintf("Queue length: %d", plexResp.QueueLength)
	if len(plexResp.Items) > 0 {
		message += "\nItems:\n" + strings.Join(plexResp.Items, "\n")
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Error: " + err.Error(),
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

	url := fmt.Sprintf("%s/list-media?type=%s&page=%d", PlexAPIURL, mediaType, page)
	if search != "" {
		url += "&search=" + url.QueryEscape(search)
	}

	resp, err := http.Get(url)
	if err != nil {
		respondError(s, i, err)
		return
	}
	defer resp.Body.Close()

	var result listMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		respondError(s, i, err)
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("%s (Page %d/%d)", strings.Title(mediaType), result.Page, result.TotalPages),
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, item := range result.Items {
		description := fmt.Sprintf("Year: %d\nRating: %.1f\nDuration: %d min",
			item.Year, item.Rating, item.Duration/60000)
		if mediaType == "shows" {
			description += fmt.Sprintf("\nSeasons: %d\nEpisodes: %d",
				item.SeasonCount, item.EpisodeCount)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   item.Title,
			Value:  description,
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

func handleRandomMedia(s *discordgo.Session, i *discordgo.InteractionCreate, mediaType string) {
	count := i.ApplicationCommandData().Options[0].IntValue()
	data := map[string]interface{}{
		"number": count,
		"type":   mediaType,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		respondError(s, i, err)
		return
	}

	resp, err := http.Post(PlexAPIURL+"/get-random-media", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		respondError(s, i, err)
		return
	}
	defer resp.Body.Close()

	var result plexResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		respondError(s, i, err)
		return
	}

	if result.Error != "" {
		respondError(s, i, fmt.Errorf(result.Error))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Random %s Queue", strings.Title(mediaType)),
		Description: fmt.Sprintf("Added %d items to queue:", len(result.Items)),
		Fields: make([]*discordgo.MessageEmbedField, len(result.Items)),
	}

	for i, title := range result.Items {
		embed.Fields[i] = &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("%d.", i+1),
			Value: title,
			Inline: true,
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}