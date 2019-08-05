package app_lifecycle

import (
	"fmt"
	"github.com/aybabtme/rgbterm"
	"github.com/isbm/spaceman/lib/app_info"
	"github.com/isbm/spaceman/lib/utils"
	"github.com/thoas/go-funk"
	"gopkg.in/urfave/cli.v1"
	"strings"
)

var Logger utils.LoggerController
var ChannelLifecycleFlags []cli.Flag

func init() {
	ChannelLifecycleFlags = []cli.Flag{
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
	ctx              *cli.Context
}

// NewChannelLifecycle constructor
func NewChannelLifecycle(context *cli.Context) *channelLifecycle {
	lifecycle := new(channelLifecycle)
	lifecycle.ctx = context

	return lifecycle
}

// Promote channel to the specific stage
func (lifecycle *channelLifecycle) promoteChannel(channelName string, init bool) string {
	currentPhase := lifecycle.extractPhaseName(channelName)
	if currentPhase == "" && !init {
		utils.Console.ExitOnStderr("Unable to get phase")
	}
	nextPhase := lifecycle.getNextPhase(currentPhase, init)

	if nextPhase != "" && !init {
		channelName = fmt.Sprintf("%s%s%s", nextPhase, lifecycle.phasesDelimiter, channelName[len(currentPhase)+1:])
	} else if nextPhase != "" && currentPhase != "" && init {
		utils.Console.ExitOnStderr("Channel is already initalised. Please just promote it.")
	} else if nextPhase != "" && currentPhase == "" && init {
		channelName = fmt.Sprintf("%s%s%s", nextPhase, lifecycle.phasesDelimiter, channelName)
	} else {
		utils.Console.ExitOnStderr("Unable to promote channel.")
	}

	return channelName
}

// Clone channel by label
func (lifecycle *channelLifecycle) CloneChannel(labelSrc string, labelDst string, details map[string]interface{}) {
	if lifecycle.ctx.String("exclude-channel") != "" {
		excludePattern := lifecycle.ctx.String("exclude-channel")
		if strings.Contains(labelSrc, excludePattern) {
			labelSrc = strings.ReplaceAll(labelSrc, excludePattern, rgbterm.FgString(excludePattern, 0xff, 0xff, 0))
			Logger.Fatal("Seems like you wanted to exclude this channel (" + labelSrc + ")?")
		}
	}
	sourceChannelLabel, exist := details["label"]
	if !exist {
		Logger.Fatal("Unable to get full data about the channel: label is missing")
	}
	cloneDetails := make(map[string]interface{})
	cloneDetails["label"] = labelDst
	cloneDetails["name"] = labelDst
	cloneDetails["summary"] = details["summary"]

	if details["parent_channel_label"] != nil {
		cloneDetails["parent_label"] = details["parent_channel_label"]
	} else {
		cloneDetails["parent_label"] = ""
	}

	Logger.Debug("Getting details about channel \"%s\"", sourceChannelLabel.(string))
	utils.RPC.RequestFuction("channel.software.clone", utils.RPC.GetSession(), sourceChannelLabel, cloneDetails, false)
}

// List available workflows
func (lifecycle *channelLifecycle) ListWorkflows(ctx *cli.Context) {
	configSections := utils.Configuration.GetConfig(ctx, "lifecycle")
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
func (lifecycle *channelLifecycle) getWorkflowConfig(name string) *map[string]interface{} {
	currentWorkflow := make(map[string]interface{})
	configSections := utils.Configuration.GetConfig(lifecycle.ctx, "lifecycle")
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

// Check if specified channel exists
func (lifecycle *channelLifecycle) GetChannelDetails(name string) map[string]interface{} {
	stuff := utils.RPC.RequestFuction("channel.software.getDetails", utils.RPC.GetSession(), name)
	return stuff.(map[string]interface{})
}

// Find what workflow currently is used and setup the phases
func (lifecycle *channelLifecycle) setCurrentWorkflow() *channelLifecycle {
	currentWorkflowName := lifecycle.ctx.String("workflow")
	if funk.Contains([]string{"", "default"}, currentWorkflowName) {
		currentWorkflowName = "default"
	}
	configuredWorkflow := lifecycle.getWorkflowConfig(currentWorkflowName)
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
	return lifecycle
}

// Set flags from CLI and configuration about current runtime session
func (lifecycle *channelLifecycle) setCurrentConfig() *channelLifecycle {
	if lifecycle.ctx.GlobalBool("quiet") && lifecycle.ctx.GlobalBool("verbose") {
		utils.Console.ExitOnUnknown("Don't know how to be quietly verbose.")
	}

	Logger = *utils.NewLoggerController(lifecycle.ctx.GlobalBool("verbose"), lifecycle.ctx.GlobalBool("verbose"),
		!lifecycle.ctx.GlobalBool("quiet"), lifecycle.ctx.GlobalBool("verbose"))
	Logger.Debug("Configuration set")

	return lifecycle
}

// Entry action for the managing channel lifecycle sub-app
func ManageChannelLifecycle(ctx *cli.Context) error {
	lifecycle := NewChannelLifecycle(ctx).setCurrentConfig().setCurrentWorkflow()

	if ctx.Bool("list-workflows") {
		lifecycle.ListWorkflows(ctx)
	} else if ctx.Bool("promote") || ctx.Bool("init") {

		channelToPromote := ctx.String("channel")
		if channelToPromote == "" {
			utils.Console.ExitOnUnknown("Channel required.")
		}
		details := lifecycle.GetChannelDetails(channelToPromote)
		promotedChannelName := lifecycle.promoteChannel(channelToPromote, ctx.Bool("init"))
		lifecycle.CloneChannel(channelToPromote, promotedChannelName, details)
		Logger.Info("Channel \"%s\" promoted to \"%s\"\n", channelToPromote, promotedChannelName)
		app_info.InfoCmd(ctx).SetCurrentConfig(ctx).ChannelDetails(promotedChannelName)
	} else {
		utils.Console.ExitOnUnknown("Don't know what to do.")
	}

	return nil
}
