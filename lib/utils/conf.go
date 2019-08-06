package utils

import (
	"fmt"
	"github.com/smallfish/simpleyaml"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
	"strings"
)

/*
	configFiles allows to keep track of existing configurations
*/
type configFiles struct {
	global  string
	local   string
	session string
	used    string
}

// Config object constructor
func NewConfig() *configFiles {
	cfg := new(configFiles)
	cfg.global = "/etc/rhn/spaceman.conf"
	cfg.local = cfg.expandPath("~/.config/spaceman/config.conf")
	cfg.session = cfg.expandPath("~/.config/spaceman/session.conf")
	cfg.used = cfg.local

	return cfg
}

// Expands "~" to "$HOME".
func (cfg *configFiles) expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}

// Return default configuration file
func (cfg *configFiles) GetDefaultConfigFile() string {
	return cfg.used
}

// Return current configuration file
func (cfg *configFiles) GetConfigFile(ctx *cli.Context) string {
	custom := ctx.GlobalString("config")
	if custom != "" {
		cfg.used = custom
	}

	return cfg.used
}

// Returns path of the session config
func (cfg *configFiles) GetSessionConfFilePath() string {
	return cfg.session
}

func (cfg *configFiles) checkFail(err error, message string) {
	if err != nil {
		log.Fatal(err)
		panic(message)
	}
}

func (cfg *configFiles) GetConfig(ctx *cli.Context, sections ...string) *map[string]interface{} {
	filename := cfg.GetConfigFile(ctx)
	if filename != "" {
		filename = cfg.expandPath(filename)
		source, err := ioutil.ReadFile(filename)
		cfg.checkFail(err, "Unable to read configuration file")

		data, err := simpleyaml.NewYaml(source)
		cfg.checkFail(err, "Unable to parse YAML data")

		content := make(map[string]interface{})
		globalConfig, err := data.Map()
		cfg.checkFail(err, "Configuration syntax error: structure expected")
		for _, section := range sections {
			sectionConfig, exist := globalConfig[section]
			if exist {
				content[section] = sectionConfig
			} else {
				log.Printf("Section '%s' does not exist", section)
			}
		}
		if len(content) == 0 {
			log.Fatal(fmt.Sprintf("No configuration found for %s sections", strings.Join(sections, ", ")))
		}

		return &content
	}
	panic("Unable to obtain configuration")
}

var Configuration configFiles

func init() {
	Configuration = *NewConfig()
}
