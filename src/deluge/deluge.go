package deluge

import (
	"fmt"
	"high-seas/src/logger"
	"high-seas/src/utils"

	"github.com/superturkey650/go-qbittorrent/qbt"
)

var user = utils.EnvVar("DELUGE_USER", "")
var password = utils.EnvVar("DELUGE_PASSWORD", "")
var ip = utils.EnvVar("DELUGE_IP", "")
var port = utils.EnvVar("DELUGE_PORT", "")

func AddTorrent(file string) error {
	url := fmt.Sprintf("http://%s:%s", ip, port)

	qb := qbt.NewClient(url)

	qb.Login(user, password)

	options := map[string]string{}

	result, err := qb.DownloadFromLink(file, options)
	if err != nil {
		logger.WriteError("Failed to add torrent.", err)
		return err
	}

	logger.WriteCMDInfo("Status Code Returned: ", result.Status)
	return nil
}
