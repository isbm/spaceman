package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/thoas/go-funk"

	"gopkg.in/urfave/cli.v1"
)

var Logger loggerController
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
	channel          string
	phases           []string
	excludedChannels []string
	phasesDelimiter  string
	verbose          bool
}

// ChannelLifecycle constructor
func ChannelLifecycle() *channelLifecycle {
	lifecycle := new(channelLifecycle)
	return lifecycle
}

// Promote channel to the specific stage
func (lifecycle *channelLifecycle) promoteChannel(channelName string, init bool) string {
	currentPhase := lifecycle.extractPhaseName(channelName)
	if currentPhase == "" && !init {
		Console.exitOnStderr("Unable to get phase")
	}
	nextPhase := lifecycle.getNextPhase(currentPhase, init)

	if nextPhase != "" && !init {
		channelName = fmt.Sprintf("%s%s%s", nextPhase, lifecycle.phasesDelimiter, channelName[len(currentPhase)+1:])
	} else if nextPhase != "" && currentPhase != "" && init {
		Console.exitOnStderr("Channel is already initalised. Please just promote it.")
	} else if nextPhase != "" && currentPhase == "" && init {
		channelName = fmt.Sprintf("%s%s%s", nextPhase, lifecycle.phasesDelimiter, channelName)
	} else {
		Console.exitOnStderr("Unable to promote channel.")
	}

	return channelName
}

// List available workflows
func (lifecycle *channelLifecycle) ListWorkflows(ctx *cli.Context) {
	configSections := configuration.getConfig(ctx, "lifecycle")
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

// Returns a phase name from the given channel name
func (lifecycle *channelLifecycle) extractPhaseName(channelName string) string {
	phase := ""
	hasPhase := false
	for _, phase := range lifecycle.phases {
		phase = phase + lifecycle.phasesDelimiter
		if len(channelName) > len(phase) && strings.HasPrefix(channelName, phase) {
			hasPhase = true
			break
		}
	}
	if hasPhase {
		// Get phase
		phase = strings.Split(channelName, lifecycle.phasesDelimiter)[0]
	}

	return phase
}

/*
Get next phase for the channel from the configured workflow.
If current phase is set to an empty string (e.g. not found in the name of the channel)
and init is set to True, then first phase of the current workflow is used.
*/
func (lifecycle *channelLifecycle) getNextPhase(currentPhase string, init bool) string {
	phase := ""
	if currentPhase == "" && init {
		phase = lifecycle.phases[0]
	} else {
		if currentPhase == lifecycle.phases[len(lifecycle.phases)-1] {
			Logger.Fatal("Error. Unable to rotate phase: reached last available already.")
		} else {
			for i, lcPhase := range lifecycle.phases {
				if lcPhase == currentPhase {
					phase = lifecycle.phases[i+1]
					break
				}
			}
		}
	}

	return phase
}

// Get workflow configuration or return default one.
func (lifecycle *channelLifecycle) getWorkflowConfig(name string, ctx *cli.Context) *map[string]interface{} {
	currentWorkflow := make(map[string]interface{})
	configSections := configuration.getConfig(ctx, "lifecycle")
	lifecycleConfig, exist := (*configSections)["lifecycle"].(map[interface{}]interface{})
	if exist {
		workflowsConfig, exist := lifecycleConfig["workflows"]
		if exist {
			workflowsData := workflowsConfig.(map[interface{}]interface{})
			for workflowName := range workflowsData {
				if workflowName == name {
					currentWorkflow[name] = workflowsData[name]
					break
				}
			}
		}
	}

	return &currentWorkflow
}

// List available channels over XML-RPC API
func (lifecycle *channelLifecycle) ListChannels(ctx *cli.Context) {
	Logger.Info("List channels")
	out := rpc.requestFuction("channel.listSoftwareChannels", rpc.session)
	tree := make(map[string][]string)

	for _, dat := range out.(Array) {
		channel := dat.(Struct)
		if channel["parent_label"] != "" {
			if !funk.Contains(tree, channel["parent_label"]) {
				tree[channel["parent_label"].(string)] = []string{}
			}
			tree[channel["parent_label"].(string)] = append(tree[channel["parent_label"].(string)], channel["label"].(string))
		} else {
			if !funk.Contains(tree, channel["label"]) {
				tree[channel["label"].(string)] = []string{}
			}
		}
	}

	if len(tree) == 0 {
		Console.exitOnStderr("No channels has been found")
	} else {
		lifecycle.outputTreeToStdout(tree)
	}
}

// Prints the tree to the STDOUT
func (lifecycle *channelLifecycle) outputTreeToStdout(tree map[string][]string) {
	rootLabelIndex := []string{}
	for label := range tree {
		rootLabelIndex = append(rootLabelIndex, label)
	}
	sort.Strings(rootLabelIndex)

	fmt.Printf("Tree of channels:\n%s\n", "\u2514\u2500\u2510")
	var branch string
	if len(rootLabelIndex) > 1 {
		branch = "\u251c"
	} else {
		branch = "\u2514"
	}
	branchSingle := branch + "\u2500\u2500"
	branchSingleEnd := "\u2514\u2500\u2500"

	for idx, label := range rootLabelIndex {
		idx++
		childLabels, exists := tree[label]
		var rootBranch string
		if idx < len(rootLabelIndex) {
			rootBranch = branchSingle
		} else {
			rootBranch = branchSingleEnd
		}
		cIdx := rgbterm.FgString(fmt.Sprintf("(%02d)", idx), 0xff, 0xff, 0) // Index of the root channel
		cLabel := rgbterm.FgString(label, 0xff, 0xff, 0xff)
		fmt.Printf("  %s%s %s\n", rootBranch, cIdx, cLabel)

		if exists && len(childLabels) > 0 {
			if idx < len(rootLabelIndex) {
				fmt.Printf("  %s ", "\u2502")
			} else {
				fmt.Printf("    ")
			}
			sort.Strings(childLabels)
			for cidx, childLabel := range childLabels {
				if cidx == 0 {
					if len(childLabels) == 1 {
						fmt.Printf("      %s %s\n", branchSingleEnd, childLabel)
					} else {
						fmt.Printf("      %s %s\n", branchSingle, childLabel)
					}
				} else if cidx < len(childLabels)-1 {
					fmt.Printf("  %s       %s %s\n", "\u2502", branchSingle, childLabel)
				} else {
					fmt.Printf("  %s       %s %s\n", "\u2502", branchSingleEnd, childLabel)
				}
			}
		}
		if idx < len(rootLabelIndex) {
			fmt.Printf("  %s\n", "\u2502")
		} else {
			fmt.Println("")
		}
	}
}

// Find what workflow currently is used and setup the phases
func (lifecycle *channelLifecycle) setCurrentWorkflow(ctx *cli.Context) {
	currentWorkflowName := ctx.String("workflow")
	if funk.Contains([]string{"", "default"}, currentWorkflowName) {
		currentWorkflowName = "default"
	}
	configuredWorkflow := lifecycle.getWorkflowConfig(currentWorkflowName, ctx)
	if len(*configuredWorkflow) == 0 {
		Logger.Debug("Using preset default workflow: \"dev\", \"uat\", \"prod\".")
		lifecycle.phases = []string{"dev", "uat", "prod"}
	} else {
		Logger.Debug("Using specified workflow: %s", currentWorkflowName)

		cfgPhases, configured := (*configuredWorkflow)[currentWorkflowName].(map[interface{}]interface{})["phases"]
		if configured && cfgPhases != nil {
			lifecycle.phases = make([]string, len(cfgPhases.([]interface{})))
			for i, v := range cfgPhases.([]interface{}) {
				lifecycle.phases[i] = v.(string)
			}
		} else {
			Logger.Fatal("Phases are not configured in this workflow, aborting.")
		}

		cfgExclude, configured := (*configuredWorkflow)[currentWorkflowName].(map[interface{}]interface{})["exclude"]
		if configured && cfgExclude != nil {
			lifecycle.excludedChannels = make([]string, len(cfgExclude.([]interface{})))
			for i, v := range cfgExclude.([]interface{}) {
				lifecycle.excludedChannels[i] = v.(string)
			}
		} else {
			Logger.Info("No channels configured to be excluded according to this workflow")
		}

		// Set delimiter
		cfgDelimiter, configured := (*configuredWorkflow)[currentWorkflowName].(map[interface{}]interface{})["delimiter"]
		if configured && cfgDelimiter != nil {
			lifecycle.phasesDelimiter = cfgDelimiter.(string)
		} else {
			lifecycle.phasesDelimiter = "-"
		}
	}
}

// Set flags from CLI and configuration about current runtime session
func (lifecycle *channelLifecycle) setCurrentConfig(ctx *cli.Context) {
	if ctx.GlobalBool("quiet") && ctx.GlobalBool("verbose") {
		Console.exitOnUnknown("Don't know how to be quietly verbose.")
	}

	Logger = *LoggerController(ctx.GlobalBool("verbose"), ctx.GlobalBool("verbose"),
		!ctx.GlobalBool("quiet"), ctx.GlobalBool("verbose"))
	Logger.Debug("Configuration set")
}

// Entry action for the managing channel lifecycle sub-app
func manageChannelLifecycle(ctx *cli.Context) error {
	lifecycle := ChannelLifecycle()
	lifecycle.setCurrentConfig(ctx)
	lifecycle.setCurrentWorkflow(ctx)

	if ctx.Bool("list-workflows") {
		lifecycle.ListWorkflows(ctx)
	} else if ctx.Bool("list-channels") {
		lifecycle.ListChannels(ctx)
	} else if ctx.Bool("promote") || ctx.Bool("init") {
		channelToPromote := ctx.String("channel")
		if channelToPromote == "" {
			Console.exitOnUnknown("Channel required.")
		}
		promotedChannelName := lifecycle.promoteChannel(channelToPromote, ctx.Bool("init"))
		Logger.Info("Channel \"%s\" promoted to \"%s\"\n", channelToPromote, promotedChannelName)
	} else {
		Console.exitOnUnknown("Don't know what to do.")
	}

	return nil
}
