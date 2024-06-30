package deluge

import (
	"high-seas/src/logger"
	"high-seas/src/utils"

	"strconv"

	"github.com/superturkey650/go-qbittorrent/qbt"
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

	qb := qbt.NewClient("http://{}:{}/", ip, numPort)

	qb.Login(user, password)

	options := &qbt.Options{}

	result, err := qb.DownloadFromLink(file, options)
	if err != nil {
		logger.WriteError("Failed to add torrent.", err)
		return err
	}

	logger.WriteCMDInfo("Results: ", result)
	return nil
}
