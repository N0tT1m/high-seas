package deluge

import (
	"fmt"
	"high-seas/src/logger"
	"high-seas/src/utils"
	"strconv"
	"strings"

	delugeclient "github.com/gdm85/go-libdeluge"
)

var (
	user     = utils.EnvVar("DELUGE_USER", "")
	password = utils.EnvVar("DELUGE_PASSWORD", "")
	ip       = utils.EnvVar("DELUGE_IP", "")
	port     = utils.EnvVar("DELUGE_PORT", "")
)

// connectToDeluge creates and connects to a deluge client
func connectToDeluge() (*delugeclient.ClientV2, error) {
	numPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("failed to convert port to number: %v", err)
	}

	deluge := delugeclient.NewV2(delugeclient.Settings{
		Hostname: ip,
		Port:     uint(numPort),
		Login:    user,
		Password: password,
	})

	logger.WriteInfo("Attempting to connect to Deluge")
	err = deluge.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to deluge: %v", err)
	}
	logger.WriteInfo("Successfully connected to Deluge")

	return deluge, nil
}

// AddTorrent adds either a magnet link or torrent URL to Deluge
func AddTorrent(file string) error {
	logger.WriteInfo(fmt.Sprintf("Initializing Deluge connection to %s:%s", ip, port))

	deluge, err := connectToDeluge()
	if err != nil {
		return err
	}

	options := &delugeclient.Options{}

	if strings.HasPrefix(file, "magnet:") {
		logger.WriteInfo(fmt.Sprintf("Sending magnet link to Deluge: %s", file))
		result, err := deluge.AddTorrentMagnet(file, options)
		if err != nil {
			return fmt.Errorf("failed to add magnet link: %v", err)
		}
		logger.WriteInfo(fmt.Sprintf("Successfully added magnet, Deluge response: %v", result))
	} else {
		logger.WriteInfo(fmt.Sprintf("Sending torrent URL to Deluge: %s", file))
		result, err := deluge.AddTorrentURL(file, options)
		if err != nil {
			return fmt.Errorf("failed to add torrent URL: %v", err)
		}
		logger.WriteInfo(fmt.Sprintf("Successfully added URL, Deluge response: %v", result))
	}

	return nil
}
