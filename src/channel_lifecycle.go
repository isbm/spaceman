package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/thoas/go-funk"

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
	channel          string
	phases           []string
	excludedChannels []string
	phasesDelimiter  string
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

// Verifies phase name if it belongs to the current workflow at all
func (lifecycle channelLifecycle) verifyPhase(phase string) {
}

// Get workflow configuration or return default one.
func (lifecycle channelLifecycle) getWorkflowConfig(name string, ctx *cli.Context) *map[string]interface{} {
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

// Find what workflow currently is used and setup the phases
func (lifecycle channelLifecycle) setCurrentWorkflow(ctx *cli.Context) {
	currentWorkflowName := ctx.String("workflow")
	if funk.Contains([]string{"", "default"}, currentWorkflowName) {
		currentWorkflowName = "default"
	}
	configuredWorkflow := lifecycle.getWorkflowConfig(currentWorkflowName, ctx)
	if len(*configuredWorkflow) == 0 {
		fmt.Println("Using preset default workflow: \"dev\", \"uat\", \"prod\".")
		lifecycle.phases = []string{"dev", "uat", "prod"}
	} else {
		fmt.Println("Using specified workflow:", currentWorkflowName)

		cfgPhases, configured := (*configuredWorkflow)[currentWorkflowName].(map[interface{}]interface{})["phases"]
		if configured {
			lifecycle.phases = make([]string, len(cfgPhases.([]interface{})))
			for i, v := range cfgPhases.([]interface{}) {
				lifecycle.phases[i] = v.(string)
			}
		} else {
			log.Fatal("Phases are not configured in this workflow, aborting.")
		}

		cfgExclude, configured := (*configuredWorkflow)[currentWorkflowName].(map[interface{}]interface{})["exclude"]
		if configured {
			lifecycle.excludedChannels = make([]string, len(cfgExclude.([]interface{})))
			for i, v := range cfgExclude.([]interface{}) {
				lifecycle.excludedChannels[i] = v.(string)
			}
		} else {
			log.Println("No channels configured to be excluded according to this workflow")
		}
	}
}

// Entry action for the managing channel lifecycle sub-app
func manageChannelLifecycle(ctx *cli.Context) error {
	lifecycle := ChannelLifecycle()
	lifecycle.setCurrentWorkflow(ctx)

	if ctx.Bool("list-workflows") {
		lifecycle.listWorkflows(ctx)
	} else if ctx.Bool("promote") {
		channelToPromote := ctx.String("channel")
		if channelToPromote == "" {
			endWithHint("Channel required.")
		}
		fmt.Println("Channel:", ctx.String("channel"))
		//lifecycle.promoteChannel(chanelName, phase)
	} else {
		endWithHint("Don't know what to do.")
	}

	return nil
}
