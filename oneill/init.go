package oneill

import (
	"github.com/rehabstudio/oneill/config"
	"github.com/rehabstudio/oneill/logger"
)

// initialise configuration, logging and a connection to the docker daemon
func Initialise() {
	config.InitConfig()
	logger.InitLogger(config.Config.LogLevel)
	InitDockerClient()
}
