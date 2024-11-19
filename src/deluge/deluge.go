package deluge

import (
	"fmt"
	"high-seas/src/logger"
	"high-seas/src/utils"

	"strconv"

	delugeclient "github.com/gdm85/go-libdeluge"
)

var user = utils.EnvVar("DELUGE_USER", "")
var password = utils.EnvVar("DELUGE_PASSWORD", "")
var ip = utils.EnvVar("DELUGE_IP", "")
var port = utils.EnvVar("DELUGE_PORT", "")

// Modify the Deluge AddTorrent function to be more verbose
func AddTorrent(file string) error {
	logger.WriteInfo(fmt.Sprintf("Initializing Deluge connection to %s:%s", ip, port))

	numPort, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("failed to convert port to number: %v", err)
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
		return fmt.Errorf("failed to connect to deluge: %v", err)
	}
	logger.WriteInfo("Successfully connected to Deluge")

	options := &delugeclient.Options{}

	logger.WriteInfo(fmt.Sprintf("Sending torrent URL to Deluge: %s", file))
	result, err := deluge.AddTorrentURL(file, options)
	if err != nil {
		return fmt.Errorf("failed to add torrent URL: %v", err)
	}

	logger.WriteInfo(fmt.Sprintf("Deluge response: %v", result))
	return nil
}
