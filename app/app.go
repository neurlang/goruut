// Package app manages application lifecycle, configuration, and shared dependencies.
package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/neurlang/goruut/helpers/log"
	"os"
	"regexp"
	"strings"
)
import "github.com/neurlang/goruut/repo/models"
import . "github.com/martinarisk/di/dependency_injection"

// App represents the application.
type App struct {
	args *Args
}

// NewApp creates a new instance of App.
func NewApp() *App {

	app := &App{}

	isSilent := os.Getenv("SILENT") != ""
	app.InitializeLogger(isSilent)

	return app
}

// LoadCmdArgs loads command-line arguments for the application.
func (app *App) LoadCmdArgs() *Args {
	app.args = &Args{}

	flag.Var(&app.args.ConfigFiles, "configfile", "Sets the config file")
	flag.Var(&app.args.ConfigDirs, "configdir", "Sets the config dir")
	flag.Parse()

	return app.args
}

// LoadConfigs loads configurations for the application.
func (app *App) LoadConfigs(_ *DependencyInjection) *Configs {

	var confs Configs

	for _, dirname := range app.args.ConfigDirs {
		files, err := os.ReadDir(dirname)
		if err != nil {
			log.Now().Fatalf("Couldn't read config dir:", err)
		}
		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".json") {
				continue
			}
			app.args.ConfigFiles = append(app.args.ConfigFiles,
				dirname+string(os.PathSeparator)+file.Name())
		}
	}
	regex_for_env_vars := regexp.MustCompile(`"\$[A-Za-z_]+`)

	for _, filename := range app.args.ConfigFiles {

		b, err := os.ReadFile(filename)
		if err != nil {
			fmt.Print(err)
			continue
		}

		// do env vars substitution from the environment
		b = []byte(regex_for_env_vars.ReplaceAllStringFunc(string(b), func(value string) string {

			envValue := os.Getenv(value[2:])

			return regex_for_env_vars.ReplaceAllString(value, `"`+envValue)
		}))

		// end vars substitution

		var conf models.AppConfig
		err = json.Unmarshal(b, &conf)
		if err != nil {
			fmt.Print(err)
			continue
		}

		confs.Configs = append(confs.Configs, conf)

		log.Field("config", filename).Infof("Loaded config")
	}

	return &confs
}

// NewServer creates a new instance of the server.
func (app *App) NewServer(di *DependencyInjection) *Server {

	conf := MustAny[*Configs](di)

	s := &Server{}

	s.Initialize(conf.GetHttpPort(), conf.GetAdminHttpPort())

	return s
}
