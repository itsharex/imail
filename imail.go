package main

import (
	"os"

	"github.com/urfave/cli"

	assets "github.com/midoks/imail/embed"
	"github.com/midoks/imail/internal/cmd"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools"
	"github.com/midoks/imail/internal/tools/syscall"
)

const Version = "0.1.0"
const AppName = "imail"

func init() {
	conf.App.Version = Version
	conf.App.Name = AppName

	conf.App.PublicFs = assets.PublicFS

}

func main() {
	app := cli.NewApp()
	app.Name = conf.App.Name
	app.Version = conf.App.Version
	app.Usage = "A simple mail service"
	app.Commands = []cli.Command{
		cmd.Service,
		cmd.Reset,
		cmd.Dkim,
		cmd.Cert,
		cmd.Check,
	}

	if err := app.Run(os.Args); err != nil {
		log.Infof("Failed to start application: %v", err)
	}
}
