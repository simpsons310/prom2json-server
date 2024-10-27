package p2jsvr

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

var usage = fmt.Sprintf(`Usage: %s --config-file CONFIG_FILE_PATH`, os.Args[0])

const defaultConfigFile = "config.yaml"

type Argument struct {
	ConfigFile string
}

func ParseArgument() (*Argument, error) {
	configFile := flag.String("config-file", defaultConfigFile, "Path to the configuration file")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	return &Argument{
		ConfigFile: *configFile,
	}, nil
}

type Config struct {
	Log     *LogConfig      `yaml:"logging"`
	Server  *ServerConfig   `yaml:"server"`
	Metrics []*MetricConfig `yaml:"metrics"`
}

func LoadConfig(path string) (*Config, error) {
	ymlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(ymlFile, &cfg); err != nil {
		return nil, err
	}

	cfg.loadDefault()

	return cfg, nil
}

func newDefaultConfig() *Config {
	return &Config{
		Log: &LogConfig{
			TimeFieldFormat: zerolog.TimeFieldFormat,
			Level:           zerolog.DebugLevel,
			JSONFormat:      false,
			WithCaller:      false,
		},
		Server: &ServerConfig{
			Port: 8080,
		},
		Metrics: []*MetricConfig{
			{
				HandlerPath:       "/metrics",
				ResponseTimeout:   10,
				URL:               "http://localhost:9090/metrics",
				TlsCert:           "",
				TlsKey:            "",
				TlsInsecureVerify: false,
			},
		},
	}
}

// Fill default config for unspecified fields
func (c *Config) loadDefault() error {
	defaultCfg := newDefaultConfig()

	if c.Log == nil {
		c.Log = defaultCfg.Log
	} else {
		if c.Log.TimeFieldFormat == "" {
			c.Log.TimeFieldFormat = defaultCfg.Log.TimeFieldFormat
		}
	}

	if c.Server == nil {
		c.Server = defaultCfg.Server
	} else {
		if c.Server.Port == 0 {
			c.Server.Port = defaultCfg.Server.Port
		}
	}

	if c.Metrics == nil {
		c.Metrics = defaultCfg.Metrics
	} else {
		for _, metric := range c.Metrics {
			if metric.HandlerPath == "" {
				metric.HandlerPath = defaultCfg.Metrics[0].HandlerPath
			}
			if metric.ResponseTimeout == 0 {
				metric.ResponseTimeout = defaultCfg.Metrics[0].ResponseTimeout
			}
			if metric.URL == "" {
				metric.URL = defaultCfg.Metrics[0].URL
			}
		}
	}

	return nil
}
