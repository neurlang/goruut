package models

import (
	"github.com/neurlang/goruut/helpers/log"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io/ioutil"
	"os"
)

// AppConfig represents the configuration for the application.
type AppConfig struct {
	Port      string
	AdminPort string

	FavIconSite string
	Logging     *struct {
		Level             string
		PrefixedFormatter *struct {
			LogFormat       string
			TimestampFormat string
		}
	}

	BuiltinDictLanguages []string
}

// GetHttpPort returns the HTTP port.
func (c *AppConfig) GetHttpPort() string {
	return c.Port
}

func (c *AppConfig) GetAdminHttpPort() string {
	return c.AdminPort
}

// GetFavIconSite returns the favorite icon site.
func (c *AppConfig) GetFavIconSite() string {
	return c.FavIconSite
}

// GetBuiltinDictLanguages returns the builtin dict languages.
func (c *AppConfig) GetBuiltinDictLanguages() (ret map[string]struct{}) {
	ret = make(map[string]struct{})
	for _, lang := range c.BuiltinDictLanguages {
		ret[lang] = struct{}{}
	}
	return
}

// ConfigureLogger configures the application's logger.
func (c *AppConfig) ConfigureLogger() {

	if c == nil {
		logrus.SetOutput(ioutil.Discard)
		return
	} else {
		logrus.SetOutput(os.Stderr)
	}

	if c.Logging == nil {
		return
	}

	// set the level
	var lvl logrus.Level
	log.Error0((&lvl).UnmarshalText([]byte(c.Logging.Level)))
	logrus.SetLevel(lvl)

	if c.Logging.PrefixedFormatter == nil {
		return
	}
	var formatter easy.Formatter

	formatter.TimestampFormat = c.Logging.PrefixedFormatter.TimestampFormat
	formatter.LogFormat = c.Logging.PrefixedFormatter.LogFormat

	log.Field("formatter", formatter).Infof("Initializing logging formatter")
	logrus.SetFormatter(&formatter)
	log.Field("formatter", formatter).Infof("Initialized logging formatter")

}
