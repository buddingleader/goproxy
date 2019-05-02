package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// ProxyConfig to start proxy service
type ProxyConfig struct {
	HTTPPort           string        `json:"httpport" mapstructure:"httpport" yaml:"httpport"`
	TCPPort            string        `json:"tcpport" mapstructure:"tcpport" yaml:"tcpport"`
	PProfPort          string        `json:"pprofport" mapstructure:"pprofport" yaml:"pprofport"`
	LBPolicy           int           `json:"lbpolicy" mapstructure:"lbpolicy" yaml:"lbpolicy"`
	RWTimeout          time.Duration `json:"rwtimeout" mapstructure:"rwtimeout" yaml:"rwtimeout"`
	PrintInterval      time.Duration `json:"printinterval" mapstructure:"printinterval" yaml:"printinterval"`
	HeartbeatKeepAlive time.Duration `json:"heartbeatkeepalive" mapstructure:"heartbeatkeepalive" yaml:"heartbeatkeepalive"`
	AliveCheckInterval time.Duration `json:"alivecheckinterval" mapstructure:"alivecheckinterval" yaml:"alivecheckinterval"`
	HandleBuffer       int           `json:"handlebuffer" mapstructure:"handlebuffer" yaml:"handlebuffer"`
}

// System configuration parameters
var (
	conf ProxyConfig
	once sync.Once
)

//
const (
	PROXYCONFPATH  = "PROXY_CONFIG_PATH"
	CONFFILENAME   = "proxy"
	GITHUBCINFOATH = "/src/github.com/wangff15386/goproxy/config"
)

// InitConfig initialize fileName.yaml configuration into viper
func InitConfig() {
	viper.SetConfigType("json")                                             // or viper.SetConfigType("JSON")
	viper.SetConfigName(CONFFILENAME)                                       // name of config file (without extension)
	viper.AddConfigPath(os.Getenv(PROXYCONFPATH))                           // path to look for the config file in
	viper.AddConfigPath(".")                                                // optionally look for config in the working directory
	viper.AddConfigPath(filepath.Join(os.Getenv("GOPATH"), GITHUBCINFOATH)) // optionally look for config in the github directory

	// Find and read the config file, handle errors reading the config file
	if err := viper.ReadInConfig(); err != nil {
		// The version of Viper we use claims the config type isn't supported when in fact the file hasn't been found
		// Display a more helpful message to avoid confusing the user.
		if strings.Contains(fmt.Sprint(err), "Unsupported Config Type") {
			log.Panicf("Could not find config file. "+
				"Please make sure that %s or current dir is set to a path which contains %s.yaml", PROXYCONFPATH, CONFFILENAME)
		}

		log.Panicln(errors.WithMessage(err, fmt.Sprintf("Error when reading %s.yaml config file", CONFFILENAME)))
	}

	if err := viper.Unmarshal(&conf); err != nil {
		log.Panicln("Error to unmarshal config, error:", err)
	}
}

// GetConfig get the system config
func GetConfig() ProxyConfig {
	once.Do(InitConfig)
	return conf
}
