package server

import (
	"flag"
	"syscall"

	"github.com/teamgram/marmota/pkg/commands"
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/config"
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	configFile = flag.String("f", "etc/ntproxy.yaml", "the config file")
)

type Server struct {
	server  *server.Server
}

func (s *Server) Initialize() error {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Infov(c)

	s.server = server.New(c)

	return nil
}

func (s *Server) RunLoop() {
	if err := s.server.Serve(); err != nil {
		logx.Errorf("run server error: %v, quit...", err)
		commands.GSignal <- syscall.SIGQUIT
	}
}

func (s *Server) Destroy() {
	s.server.Close()
}
