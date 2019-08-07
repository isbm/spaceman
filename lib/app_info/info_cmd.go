package app_info

import (
	"fmt"
	"github.com/aybabtme/rgbterm"
	"github.com/isbm/go-asciitable"
	"github.com/isbm/spaceman/lib/outputters"
	"github.com/isbm/spaceman/lib/utils"
	"github.com/thoas/go-funk"
	"github.com/urfave/cli"
	"sort"
)

var Logger utils.LoggerController
var InfoCmdFlags []cli.Flag

func init() {
	InfoCmdFlags = []cli.Flag{
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
	ctx     *cli.Context
}

// ChannelLifecycle constructor
func NewInfoCmd(ctx *cli.Context) *infoCmd {
	nfo := new(infoCmd)
	nfo.ctx = ctx
	return nfo
}

func (nfo *infoCmd) ChannelDetails(channel string) {
	if channel == "" {
		channel = nfo.ctx.String("channel")
	}
	out := utils.RPC.RequestFuction("channel.software.getDetails", utils.RPC.GetSession(), channel)

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
	out := utils.RPC.RequestFuction("channel.listSoftwareChannels", utils.RPC.GetSession())
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
		utils.Console.ExitOnStderr("No channels has been found")
	} else {
		outputters.NewAnsiCLI().Tree(tree)
	}
}

// Print channel info to the STDOUT
func (nfo *infoCmd) printMapInfo(data map[string]interface{}) map[string]interface{} {
	unprocessedData := make(map[string]interface{})
	activeLabelMaker := utils.NewLabels(true, 0xff, 0xff, 0)
	passiveLabelMaker := utils.NewLabels(true, 0x80, 0x80, 0x80)

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
				name = passiveLabelMaker.MapKeyToLabel(name)
			} else {
				name = activeLabelMaker.MapKeyToLabel(name)
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
func (nfo *infoCmd) SetCurrentConfig() *infoCmd {
	if nfo.ctx.GlobalBool("quiet") && nfo.ctx.GlobalBool("verbose") {
		utils.Console.ExitOnUnknown("Don't know how to be quietly verbose.")
	}

	Logger = *utils.NewLoggerController(nfo.ctx.GlobalBool("verbose"), nfo.ctx.GlobalBool("verbose"),
		!nfo.ctx.GlobalBool("quiet"), nfo.ctx.GlobalBool("verbose"))
	Logger.Debug("Configuration set")

	return nfo
}

// Entry action for the info sub-app
func MainInfoCmd(ctx *cli.Context) error {
	nfo := NewInfoCmd(ctx).SetCurrentConfig()
	utils.RPC.Connect((*utils.Configuration.GetConfig(ctx, "server")))
	if ctx.String("channel") != "" {
		nfo.ChannelDetails("")
	} else if ctx.Bool("list-channels") {
		nfo.ListAvailableChannels()
	} else {
		utils.Console.ExitOnUnknown("Don't know what kind of info you would like to have.")
	}
	return nil
}
