package main

import (
	"log"
	"os"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "Channel Tools"
	app.Usage = "Organise Uyuni channels"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: configuration.used,
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
			Action:  manageChannelLifecycle,
			Flags:   channelLifecycleFlags,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
