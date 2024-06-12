package app

import (
	"github.com/neurlang/goruut/repo/models"
)

import "net"
import "net/http"

// StringArgs represents string command-line arguments.
type StringArgs []string

// String returns the string value of StringArgs.
func (s *StringArgs) String() string {
	return ""
}

// Set sets the value of StringArgs.
func (s *StringArgs) Set(s2 string) error {
	*s = append(*s, s2)
	return nil
}

// Args represents command-line arguments.
type Args struct {
	ConfigDirs  StringArgs
	ConfigFiles StringArgs
}

// Configs represents application configurations.
type Configs struct {
	Configs []models.AppConfig
}

// InitializeLogger initializes the logger for the application.
func (*App) InitializeLogger(isSilent bool) {
	if isSilent {
		(*models.AppConfig)(nil).ConfigureLogger()
	}

}

// ConfigureLogger configures the logger for the application.
func (ac *Configs) ConfigureLogger() {

	for _, conf := range ac.Configs {
		conf.ConfigureLogger()
	}
}

// Controllers represents application controllers.
type Controllers struct {
	httpServerHandler http.Handler

	adminServerHandler http.Handler
}

// Server represents the application server.
type Server struct {
	httpServerListener net.Listener
	httpServer         *http.Server
	httpServerHandler  http.Handler

	adminServerListener net.Listener
	adminServer         *http.Server
	adminServerHandler  http.Handler
}
