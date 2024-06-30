package deluge

import (
	"high-seas/src/logger"
	"high-seas/src/utils"

	"strconv"
)

var user = utils.EnvVar("DELUGE_USER", "")
var password = utils.EnvVar("DELUGE_PASSWORD", "")
var ip = utils.EnvVar("DELUGE_IP", "")
var port = utils.EnvVar("DELUGE_PORT", "")

func AddTorrent(file string) error {
	numPort, err := strconv.Atoi(port)
	if err != nil {
		logger.WriteError("Failed to convert port to a number.", err)
	}

	deluge := delugeclient.NewV2(delugeclient.Settings{
		Hostname: ip,
		Port:     uint(numPort),
		Login:    user,
		Password: password,
	})

	err = deluge.Connect()
	if err != nil {
		logger.WriteError("Could not connect to deluge.", err)
	}

	options := &delugeclient.Options{}

	result, err := deluge.AddTorrentURL(file, options)
	if err != nil {
		logger.WriteError("Failed to add torrent.", err)
		return err
	}

	logger.WriteCMDInfo("Results: ", result)
	return nil
}
