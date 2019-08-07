package app_lifecycle

import (
	"fmt"
	"github.com/aybabtme/rgbterm"
	"github.com/isbm/spaceman/lib/app_info"
	"github.com/isbm/spaceman/lib/utils"
	"github.com/thoas/go-funk"
	"github.com/urfave/cli"
	"strings"
	"time"
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
		cli.BoolFlag{
			Name:   "m, merge",
			Usage:  "merge to the existing channel, if it already exists",
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
	phases           []string
	excludedChannels []string
	filterChannels   []string
	phasesDelimiter  string
	ctx              *cli.Context
}

// NewChannelLifecycle constructor
func NewChannelLifecycle(context *cli.Context) *channelLifecycle {
	lifecycle := new(channelLifecycle)
	lifecycle.ctx = context
	lifecycle.phasesDelimiter = "-"

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

// Check if the destination channel already exists and thus needs a merger instead of new cloning.
func (lifecycle *channelLifecycle) needsMerge(labelDst string) bool {
	needs := false
	for _, channelData := range utils.RPC.RequestFuction("channel.listSoftwareChannels", utils.RPC.GetSession()).([]interface{}) {
		label := channelData.(map[string]interface{})["label"].(string)
		if label == labelDst {
			needs = true
		}
	}

	return needs
}

// Tell if the channel is filtered-out
func (lifecycle *channelLifecycle) isFiltered(label string) string {
	prefix := ""
	for _, channelPrefix := range lifecycle.filterChannels {
		if strings.HasPrefix(label, channelPrefix) {
			prefix = channelPrefix
			break
		}
	}
	return prefix
}

// Tell if the channel is excluded
func (lifecycle *channelLifecycle) isExcluded(label string) string {
	exclusion := ""
	for _, excl := range lifecycle.excludedChannels {
		if strings.Contains(label, excl) {
			exclusion = excl
			break
		}
	}
	return exclusion
}

// Merge channels
func (lifecycle *channelLifecycle) MergeChannels(labelSrc string, labelDst string) {
	if prefix := lifecycle.isFiltered(labelDst); prefix != "" {
		utils.Console.ExitOnStderr(fmt.Sprintf("Channel \"%s\" is filtered-out in this workflow.",
			strings.Replace(labelSrc, prefix, rgbterm.FgString(prefix, 0xff, 0xff, 0), 1)))
	} else if excl := lifecycle.isExcluded(labelSrc); excl != "" {
		utils.Console.ExitOnStderr(fmt.Sprintf("Channel \"%s\" is marked as excluded by this workflow.",
			strings.ReplaceAll(labelSrc, excl, rgbterm.FgString(excl, 0xff, 0xff, 0))))
	}
	if lifecycle.ctx.Bool("clear-channel") || lifecycle.ctx.Bool("rollback") {
		lifecycle.ClearChannel(labelDst)
	}
	Logger.Info("Merging errata from channel \"%s\" to channel \"%s\"", labelSrc, labelDst)
	utils.RPC.RequestFuction("channel.software.mergeErrata", utils.RPC.GetSession(), labelSrc, labelDst)
	//Logger.Info("Added %i packages", len(packageList))

	Logger.Info("Merging packages from channel \"%s\" to channel \"%s\"", labelSrc, labelDst)
	utils.RPC.RequestFuction("channel.software.mergePackages", utils.RPC.GetSession(), labelSrc, labelDst)
}

// Clears all the errata in this channel
func (lifecycle *channelLifecycle) ClearChannel(label string) {
	Logger.Debug("Clear all errata from \"%s\"", label)
	buff := make([]string, 0)
	for _, errata := range utils.RPC.RequestFuction("channel.software.listErrata", utils.RPC.GetSession(), label).([]interface{}) {
		buff = append(buff, errata.(map[string]interface{})["advisory_name"].(string))
	}
	utils.RPC.RequestFuction("channel.software.removeErrata", utils.RPC.GetSession(), label, buff, false)
	buff = nil

	Logger.Debug("Remove all packages from \"%s\"", label)
	for _, pkg := range utils.RPC.RequestFuction("channel.software.listAllPackages", utils.RPC.GetSession(), label).([]interface{}) {
		buff = append(buff, pkg.(map[string]interface{})["id"].(string))
	}
	utils.RPC.RequestFuction("channel.software.removePackages", utils.RPC.GetSession(), label, buff)

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

// MakeArchiveLabel creates label with "archive-YYYYMMDD" prefix
func (lifecycle *channelLifecycle) MakeArchiveLabel(labelSrc string) (string, error) {
	var err error
	var label string
	if strings.HasPrefix(labelSrc, "archive-") {
		err = fmt.Errorf("The channel \"%s\" seems already archived", labelSrc)
	} else {
		current := time.Now()
		label = fmt.Sprintf("archive-%d%02d%02d-%s", current.Year(), current.Month(), current.Day(), labelSrc)
	}
	return label, err
}

// List available workflows
func (lifecycle *channelLifecycle) ListWorkflows() {
	configSections := utils.Configuration.GetConfig(lifecycle.ctx, "lifecycle")
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
			Logger.Info("No channels configured to be excluded, according to this workflow")
		}

		cfgFilter, configured := (*configuredWorkflow)[currentWorkflowName].(map[interface{}]interface{})["filter"]
		if configured && cfgFilter != nil {
			lifecycle.filterChannels = make([]string, len(cfgFilter.([]interface{})))
			for i, v := range cfgFilter.([]interface{}) {
				lifecycle.filterChannels[i] = v.(string)
			}
		} else {
			Logger.Info("No channels configured to be filtered by prefix, according to this workflow")
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
	utils.RPC.Connect((*utils.Configuration.GetConfig(ctx, "server")))

	if ctx.Bool("promote") || ctx.Bool("init") || ctx.Bool("archive") {
		if ctx.String("channel") == "" {
			utils.Console.ExitOnUnknown("Channel required.")
		}
	}

	if ctx.Bool("list-workflows") {
		lifecycle.ListWorkflows()
	} else if ctx.Bool("promote") || ctx.Bool("init") {
		channelToPromote := ctx.String("channel")
		details := lifecycle.GetChannelDetails(channelToPromote)
		promotedChannelName := lifecycle.promoteChannel(channelToPromote, ctx.Bool("init"))

		if lifecycle.needsMerge(promotedChannelName) {
			lifecycle.MergeChannels(channelToPromote, promotedChannelName)
		} else {
			lifecycle.CloneChannel(channelToPromote, promotedChannelName, details)
		}

		Logger.Info("Channel \"%s\" promoted to \"%s\"\n", channelToPromote, promotedChannelName)
		app_info.NewInfoCmd(ctx).SetCurrentConfig().ChannelDetails(promotedChannelName)
	} else {
		utils.Console.ExitOnUnknown("Don't know what to do.")
	}

	return nil
}
