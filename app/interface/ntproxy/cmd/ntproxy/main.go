package main

import (
	"github.com/devops-ntpro/teamgram-server/app/interface/ntproxy/internal/server/server"

	"github.com/teamgram/marmota/pkg/commands"
)

func main() {
	commands.Run(new(server.Server))
}
