package main

import (
	"fmt"
	"log"

	"gopkg.in/urfave/cli.v1"
)

var channelLifecycleFlags []cli.Flag

func init() {
	channelLifecycleFlags = []cli.Flag{
		cli.BoolFlag{
			Name:   "init",
			Usage:  "initialise a development channel",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "promote",
			Usage:  "promote a channel to the next phase",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "archive",
			Usage:  "archive a channel",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "rollback",
			Usage:  "rollback last operation",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "l, list-channels",
			Usage:  "list existing channels",
			Hidden: false,
		},
		cli.StringFlag{
			Name:  "c, channel",
			Usage: "use configured worflowdelimiter used between workflow and channel namechannel to init/promote/archive/rollback",
		},
		cli.BoolFlag{
			Name:   "C, clear-channel",
			Usage:  "clear all packages/errata from the channel before merging",
			Hidden: false,
		},
		cli.StringFlag{
			Name:  "x, exclude-channel",
			Usage: "skip specified channels",
		},
		cli.BoolFlag{
			Name:   "e, no-children",
			Usage:  "skip all child channels",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "r, no-errata",
			Usage:  "don't merge errata data when promoting a channel",
			Hidden: false,
		},
		cli.StringFlag{
			Name:  "P, phases",
			Usage: "comma-separated list of phases",
			Value: "dev,test,prod",
		},
		cli.StringFlag{
			Name:  "u, username, user",
			Usage: "Uyuni username",
		},
		cli.StringFlag{
			Name:  "p, password, pwd",
			Usage: "Uyuni password",
		},
		cli.StringFlag{
			Name:  "s, server",
			Usage: "Uyuni server",
		},
		cli.BoolFlag{
			Name:   "n, dry-run",
			Usage:  "don't perform any real operations",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "t, tolerant",
			Usage:  "be tolerant of errors",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "d, debug",
			Usage:  "verbose mode",
			Hidden: false,
		},
		cli.StringFlag{
			Name:  "w, workflow",
			Usage: "use configured worflow",
		},
		cli.StringFlag{
			Name:  "D, delimiter",
			Usage: "use configured worflowdelimiter used between workflow and channel name",
		},
		cli.BoolFlag{
			Name:   "f, list-workflows",
			Usage:  "list configured workflows",
			Hidden: false,
		},
	}
}

type channelLifecycle struct {
	channel string
}

// ChannelLifecycle constructor
func ChannelLifecycle() *channelLifecycle {
	lifecycle := new(channelLifecycle)
	return lifecycle
}

// List available workflows
func (lifecycle channelLifecycle) listWorkflows(ctx *cli.Context) {
	configSections := configuration.getConfig(ctx, "lifecycle")
	if len(*configSections) == 0 {
		log.Fatal("No lifecycle configuration found")
	}
	lifecycleConfig, exist := (*configSections)["lifecycle"].(map[interface{}]interface{})
	if exist {
		workflowsConfig, exist := lifecycleConfig["workflows"]
		if exist {
			idx := 1
			workflowsData := workflowsConfig.(map[interface{}]interface{})
			if len(workflowsData) > 1 {
				fmt.Println("Configured additional workflows:")
				for workflowName := range workflowsData {
					if workflowName != "default" {
						fmt.Printf("  %d. %s\n", idx, workflowName)
						idx++
					}
				}
				fmt.Println()
			} else {
				fmt.Println("No additional workflows configured")
			}
		}
	}
}

// Entry action for the managing channel lifecycle sub-app
func manageChannelLifecycle(ctx *cli.Context) error {
	lifecycle := ChannelLifecycle()
	if ctx.Bool("list-workflows") {
		lifecycle.listWorkflows(ctx)
	} else {
		fmt.Println("Don't know what to do. Try --help for more details, perhaps?")
	}

	return nil
}
