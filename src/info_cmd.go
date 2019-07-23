package main

import (
	"fmt"
	"sort"

	"github.com/aybabtme/rgbterm"

	"github.com/gosuri/uitable"

	"gopkg.in/urfave/cli.v1"
)

var infoCmdFlags []cli.Flag

func init() {
	infoCmdFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "c, channel",
			Usage: "Get information about specified channel",
		},
	}
}

type infoCmd struct {
	verbose bool
	cliArgs *cli.Context
}

// ChannelLifecycle constructor
func InfoCmd(ctx *cli.Context) *infoCmd {
	nfo := new(infoCmd)
	nfo.cliArgs = ctx
	return nfo
}

func (nfo *infoCmd) ChannelDetails() {
	channel := nfo.cliArgs.String("channel")
	out := rpc.requestFuction("channel.software.getDetails", rpc.session, channel)

	fmt.Printf("\nDetails of channel \"%s\":\n", channel)
	contentSources := nfo.printMapInfo(out.(map[string]interface{}))

	// Content sources
	for _, udContent := range contentSources {
		for _, contentSource := range udContent.([]interface{}) {
			fmt.Printf("\nContent source:\n")
			nfo.printMapInfo(contentSource.(map[string]interface{}))
		}
	}

}

// Print channel info to the STDOUT
func (nfo *infoCmd) printMapInfo(data map[string]interface{}) map[string]interface{} {
	unprocessedData := make(map[string]interface{}, 0)

	table := uitable.New()
	table.MaxColWidth = 80
	table.Separator = "  "
	table.Wrap = false
	table.AddRow(rgbterm.FgString("NAME", 0xff, 0xff, 0xff), rgbterm.FgString("DESCRIPTION", 0xff, 0xff, 0xff))

	activeLabelMaker := NewLabels(true, 0xff, 0xff, 0)
	passiveLabelMaker := NewLabels(true, 0x80, 0x80, 0x80)

	dataNames := make([]string, len(data))
	idx := 0
	for name := range data {
		dataNames[idx] = name
		idx++
	}
	sort.Strings(dataNames)

	for _, name := range dataNames {
		descr := data[name]

		switch descr.(type) {
		case []interface{}:
			unprocessedData[name] = descr
		default:
			if descr == nil {
				descr = rgbterm.FgString("n/a", 0x80, 0x80, 0x80)
				name = passiveLabelMaker.mapKeyToLabel(name)
			} else {
				name = activeLabelMaker.mapKeyToLabel(name)
			}
			table.AddRow(name, descr)
		}
	}
	fmt.Println(table)
	fmt.Println()

	return unprocessedData
}

// Set flags from CLI and configuration about current runtime session
func (nfo *infoCmd) setCurrentConfig(ctx *cli.Context) {
	if ctx.GlobalBool("quiet") && ctx.GlobalBool("verbose") {
		Console.exitOnUnknown("Don't know how to be quietly verbose.")
	}

	Logger = *LoggerController(ctx.GlobalBool("verbose"), ctx.GlobalBool("verbose"),
		!ctx.GlobalBool("quiet"), ctx.GlobalBool("verbose"))
	Logger.Debug("Configuration set")
}

// Entry action for the info sub-app
func mainInfoCmd(ctx *cli.Context) error {
	nfo := InfoCmd(ctx)
	nfo.setCurrentConfig(ctx)
	if ctx.String("channel") != "" {
		nfo.ChannelDetails()
	} else {
		Console.exitOnUnknown("Don't know what kind of info you would like to have.")
	}
	return nil
}
