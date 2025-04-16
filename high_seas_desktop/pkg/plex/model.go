// pkg/plex/models.go

package plex

// Library represents a Plex library section
type Library struct {
	Key       string     `xml:"key,attr" json:"key"`
	Type      string     `xml:"type,attr" json:"type"`
	Title     string     `xml:"title,attr" json:"title"`
	Agent     string     `xml:"agent,attr" json:"agent"`
	Scanner   string     `xml:"scanner,attr" json:"scanner"`
	Language  string     `xml:"language,attr" json:"language"`
	UUID      string     `xml:"uuid,attr" json:"uuid"`
	UpdatedAt int64      `xml:"updatedAt,attr" json:"updatedAt"`
	CreatedAt int64      `xml:"createdAt,attr" json:"createdAt"`
	Location  string     `xml:"Location>path,attr" json:"location"`
	Sections  []*Library `xml:"-" json:"-"`
}

// MediaItem represents a Plex media item (movie, TV show, music, etc.)
type MediaItem struct {
	Key           string  `json:"key"`
	Title         string  `json:"title"`
	Type          string  `json:"type"`
	Summary       string  `json:"summary"`
	Year          int     `json:"year"`
	Thumb         string  `json:"thumb"`
	Art           string  `json:"art"`
	Duration      int     `json:"duration"`
	AddedAt       int64   `json:"addedAt"`
	UpdatedAt     int64   `json:"updatedAt"`
	ViewCount     int     `json:"viewCount"`
	ViewOffset    int     `json:"viewOffset"`
	LastViewedAt  int64   `json:"lastViewedAt"`
	OriginalTitle string  `json:"originalTitle"`
	TitleSort     string  `json:"titleSort"`
	ContentRating string  `json:"contentRating"`
	Rating        float64 `json:"rating"`
	Studio        string  `json:"studio"`
	Tagline       string  `json:"tagline"`
	Media         []Media `json:"Media"`
	Genres        []Tag   `json:"Genre"`
	Directors     []Tag   `json:"Director"`
	Writers       []Tag   `json:"Writer"`
	Roles         []Role  `json:"Role"`

	// TV show specific fields
	Index            int    `json:"index"`
	ParentIndex      int    `json:"parentIndex"`
	ParentTitle      string `json:"parentTitle"`
	GrandparentKey   string `json:"grandparentKey"`
	GrandparentTitle string `json:"grandparentTitle"`
	GrandparentThumb string `json:"grandparentThumb"`

	// Music specific fields
	ParentKey   string `json:"parentKey"`
	ArtistTitle string `json:"artistTitle"`
	AlbumTitle  string `json:"albumTitle"`
}

// Media represents a media container (video, audio, etc.)
type Media struct {
	ID              int     `json:"id"`
	Duration        int     `json:"duration"`
	Bitrate         int     `json:"bitrate"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	AspectRatio     float64 `json:"aspectRatio"`
	AudioChannels   int     `json:"audioChannels"`
	AudioCodec      string  `json:"audioCodec"`
	VideoCodec      string  `json:"videoCodec"`
	VideoResolution string  `json:"videoResolution"`
	Container       string  `json:"container"`
	VideoFrameRate  string  `json:"videoFrameRate"`
	Part            []Part  `json:"Part"`
}

// Part represents a media part
type Part struct {
	ID        int      `json:"id"`
	Key       string   `json:"key"`
	Duration  int      `json:"duration"`
	File      string   `json:"file"`
	Size      int64    `json:"size"`
	Container string   `json:"container"`
	Streams   []Stream `json:"Stream"`
}

// Stream represents a media stream
type Stream struct {
	ID                int    `json:"id"`
	StreamType        int    `json:"streamType"`
	Codec             string `json:"codec"`
	Index             int    `json:"index"`
	Bitrate           int    `json:"bitrate"`
	BitDepth          int    `json:"bitDepth"`
	ChromaLocation    string `json:"chromaLocation"`
	ChromaSubsampling string `json:"chromaSubsampling"`
	CodedHeight       int    `json:"codedHeight"`
	CodedWidth        int    `json:"codedWidth"`
	ColorPrimaries    string `json:"colorPrimaries"`
	ColorRange        string `json:"colorRange"`
	ColorSpace        string `json:"colorSpace"`
	ColorTrc          string `json:"colorTrc"`
	Default           bool   `json:"default"`
	DisplayTitle      string `json:"displayTitle"`
	ExtDisplayTitle   string `json:"extDisplayTitle"`
	Height            int    `json:"height"`
	Width             int    `json:"width"`
	Language          string `json:"language"`
	LanguageCode      string `json:"languageCode"`
	Selected          bool   `json:"selected"`
	Title             string `json:"title"`
}

// Tag represents a generic tag (genre, director, etc.)
type Tag struct {
	ID     int    `json:"id"`
	Filter string `json:"filter"`
	Tag    string `json:"tag"`
	TagKey string `json:"tagKey"`
	Thumb  string `json:"thumb"`
}

// Role represents a cast/crew member
type Role struct {
	ID     int    `json:"id"`
	Filter string `json:"filter"`
	Tag    string `json:"tag"`
	TagKey string `json:"tagKey"`
	Role   string `json:"role"`
	Thumb  string `json:"thumb"`
}

// Server represents a Plex Media Server
type Server struct {
	Name    string
	URL     string
	Address string
	Port    string
}
