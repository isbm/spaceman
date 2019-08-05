package main

import (
	"fmt"
	"github.com/aybabtme/rgbterm"
	"github.com/isbm/asciitable"
	"github.com/thoas/go-funk"
	"gopkg.in/urfave/cli.v1"
	"sort"
)

var infoCmdFlags []cli.Flag

func init() {
	infoCmdFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "c, channel",
			Usage: "Get information about specified channel",
		},
		cli.BoolFlag{
			Name:   "l, list-channels",
			Usage:  "list existing channels",
			Hidden: false,
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

func (nfo *infoCmd) ChannelDetails(channel string) {
	if channel == "" {
		channel = nfo.cliArgs.String("channel")
	}
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

// List available channels tree
func (nfo *infoCmd) ListAvailableChannels() {
	Logger.Info("List channels")
	out := rpc.requestFuction("channel.listSoftwareChannels", rpc.session)
	tree := make(map[string][]string)

	for _, dat := range out.([]interface{}) {
		channel := dat.(map[string]interface{})
		if channel["parent_label"] != nil {
			if !funk.Contains(tree, channel["parent_label"]) {
				tree[channel["parent_label"].(string)] = []string{}
			}
			tree[channel["parent_label"].(string)] = append(tree[channel["parent_label"].(string)], channel["label"].(string))
		} else {
			if channel["label"] != nil && !funk.Contains(tree, channel["label"]) {
				tree[channel["label"].(string)] = []string{}
			}
		}
	}

	if len(tree) == 0 {
		Console.exitOnStderr("No channels has been found")
	} else {
		NewAnsiCLI().Tree(tree)
	}
}

// Print channel info to the STDOUT
func (nfo *infoCmd) printMapInfo(data map[string]interface{}) map[string]interface{} {
	unprocessedData := make(map[string]interface{})
	activeLabelMaker := NewLabels(true, 0xff, 0xff, 0)
	passiveLabelMaker := NewLabels(true, 0x80, 0x80, 0x80)

	dataNames := make([]string, len(data))
	idx := 0
	for name := range data {
		dataNames[idx] = name
		idx++
	}
	sort.Strings(dataNames)

	tableDataContainer := asciitable.NewTableData().SetHeader(rgbterm.FgString("NAME", 0xff, 0xff, 0xff),
		rgbterm.FgString("DESCRIPTION", 0xff, 0xff, 0xff))

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

			tableDataContainer.AddRow(name, descr)
		}
	}

	tableStyle := asciitable.NewBorderStyle(asciitable.BORDER_SINGLE_THIN, asciitable.BORDER_SINGLE_THIN).
		SetBorderVisible(false).
		SetGridVisible(false).
		SetHeaderVisible(true).
		SetHeaderStyle(asciitable.BORDER_SINGLE_THICK).
		SetTableWidthFull(true)

	table := asciitable.NewSimpleTable(tableDataContainer, tableStyle).
		SetCellPadding(1).
		SetTextWrap(true).
		SetColWidth(25, -1).
		SetColAlign(asciitable.ALIGN_RIGHT, 0).
		SetColTextWrap(false, 0)

	fmt.Println(table.Render())
	fmt.Println()

	return unprocessedData
}

// Set flags from CLI and configuration about current runtime session
func (nfo *infoCmd) setCurrentConfig(ctx *cli.Context) *infoCmd {
	if ctx.GlobalBool("quiet") && ctx.GlobalBool("verbose") {
		Console.exitOnUnknown("Don't know how to be quietly verbose.")
	}

	Logger = *LoggerController(ctx.GlobalBool("verbose"), ctx.GlobalBool("verbose"),
		!ctx.GlobalBool("quiet"), ctx.GlobalBool("verbose"))
	Logger.Debug("Configuration set")

	return nfo
}

// Entry action for the info sub-app
func mainInfoCmd(ctx *cli.Context) error {
	nfo := InfoCmd(ctx).setCurrentConfig(ctx)
	if ctx.String("channel") != "" {
		nfo.ChannelDetails("")
	} else if ctx.Bool("list-channels") {
		nfo.ListAvailableChannels()
	} else {
		Console.exitOnUnknown("Don't know what kind of info you would like to have.")
	}
	return nil
}
