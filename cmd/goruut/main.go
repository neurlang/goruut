// Main is the main production package for the application executable
package main

import (
	"github.com/martinarisk/di/dependency_injection"
)
import application "github.com/neurlang/goruut/app"
import "github.com/neurlang/goruut/dicts"
import "github.com/neurlang/goruut/loader"
import "github.com/neurlang/goruut/repo/interfaces"

// main is the main function for the application executable
func main() {

	var di = dependency_injection.NewDependencyInjection()

	di.Add((interfaces.DictGetter)(dicts.DictGetter{}))

	var app = application.NewApp()

	di.Add(app.LoadCmdArgs())

	var conf = app.LoadConfigs(di)

	conf.ConfigureLogger()

	di.Add((interfaces.LoadModels)(conf))

	var loader = loader.NewLoader(di)

	di.Add((interfaces.DictGetter)(loader))
	di.Add((interfaces.IpaFlavor)(conf))
	di.Add((interfaces.PolicyMaxWords)(conf))

	di.Add(conf)

	var server = app.NewServer(di)

	di.Add(server)

	di.Add(app.NewAppViews(di))

	di.Add(app.NewAppControllers(di))

	server.RunForever()
}
