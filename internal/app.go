package p2jsvr

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type contextKey string

type App struct {
	logger  zerolog.Logger
	config  *Config
	metrics []*Metric
}

func NewApp(cfg *Config) (*App, error) {
	metrics := make([]*Metric, 0)
	for _, m := range cfg.Metrics {
		metric, err := NewMetric(m)
		if err != nil {
			panic(err)
		}
		metrics = append(metrics, metric)
	}

	return &App{
		config:  cfg,
		logger:  NewLogger(cfg.Log),
		metrics: metrics,
	}, nil
}

func (a *App) RegisterHandler(mux *http.ServeMux) {
	logger := a.logger
	for _, metric := range a.metrics {
		logger.Info().
			Str("path", metric.config.HandlerPath).
			Str("url", metric.URL()).
			Msg("register handler")

		mux.HandleFunc(metric.config.HandlerPath, metric.MetricHandler)
	}
}

func (a *App) ContextWithLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, a.logger)
}

func LoggerFromContext(ctx context.Context) zerolog.Logger {
	logger, ok := ctx.Value(loggerKey).(zerolog.Logger)
	if !ok {
		return log.Logger
	}
	return logger
}
