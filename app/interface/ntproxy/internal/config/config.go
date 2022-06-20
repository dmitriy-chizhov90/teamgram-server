package config

import (
	"github.com/teamgram/marmota/pkg/net2"
)

type Config struct {
	MaxProc        int
	ServiceId      int
	Server         *net2.TcpServerConfig
	Client         *net2.ClientConfig
}
