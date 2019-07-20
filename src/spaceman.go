package main

import (
	"log"
	"os"

	"gopkg.in/urfave/cli.v1"
)

var flags []cli.Flag

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
