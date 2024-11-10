package p2jsvr

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const loggerKey contextKey = "logger"

type LogConfig struct {
	TimeFieldFormat string        `yaml:"time_format"`
	Level           zerolog.Level `yaml:level"`
	JSONFormat      bool          `yaml:"json_format"`
	WithCaller      bool          `yaml:"with_caller"`
}

func NewLogger(cfg *LogConfig) zerolog.Logger {
	zerolog.TimeFieldFormat = cfg.TimeFieldFormat
	zerolog.SetGlobalLevel(cfg.Level)
	if !cfg.JSONFormat {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	if cfg.WithCaller {
		log.Logger = log.Logger.With().Caller().Logger()
	}

	return log.Logger
}
