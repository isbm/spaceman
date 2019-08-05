package main

import (
	"github.com/isbm/spaceman/lib/app_info"
	"github.com/isbm/spaceman/lib/app_lifecycle"
	"github.com/isbm/spaceman/lib/utils"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "Channel Tools"
	app.Usage = "Organise Uyuni channels"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: utils.Configuration.GetDefaultConfigFile(),
			Usage: "Configuration file",
		},
		cli.BoolFlag{
			Name:   "V, verbose",
			Usage:  "Print log messages about every step",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "Q, quiet",
			Usage:  "Turn off entire logging (no errors either), only standard messages, if any",
			Hidden: false,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "lifecycle",
			Aliases: []string{"lc"},
			Usage:   "Manage channel lifecycle",
			Action:  app_lifecycle.ManageChannelLifecycle,
			Flags:   app_lifecycle.ChannelLifecycleFlags,
		},
		{
			Name:    "info",
			Aliases: []string{"in"},
			Usage:   "Information about channels, packages, machines etc.",
			Action:  app_info.MainInfoCmd,
			Flags:   app_info.InfoCmdFlags,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
