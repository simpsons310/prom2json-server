package p2jsvr

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
)

type MetricConfig struct {
	HandlerPath       string `yaml:"handler_path"`
	URL               string `yaml:"url"`
	ResponseTimeout   int    `yaml:"response_timeout"`
	TlsCert           string `yaml:"tls_cert"`
	TlsKey            string `yaml:"tls_key"`
	TlsInsecureVerify bool   `yaml:"tls_insecure_verify"`
}

type Metric struct {
	url       *url.URL
	config    *MetricConfig
	transport *http.Transport
}

func NewMetric(cfg *MetricConfig) (*Metric, error) {
	// parse url
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}

	// make transport
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = time.Minute
	tlsConfig := &tls.Config{InsecureSkipVerify: cfg.TlsInsecureVerify}
	if cfg.TlsCert != "" && cfg.TlsKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TlsCert, cfg.TlsKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	transport.TLSClientConfig = tlsConfig

	// new struct
	return &Metric{
		url:       u,
		config:    cfg,
		transport: transport,
	}, nil
}

func (m *Metric) URL() string {
	return m.url.String()
}

func (m *Metric) MetricHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := LoggerFromContext(ctx)
	w.Header().Set("Content-Type", "application/json")

	logger.Info().
		Str("host", r.Host).
		Str("tag", m.url.String()).
		Msg("fetch metrics")

	mfChan := make(chan *dto.MetricFamily, 1024)
	if err := prom2json.FetchMetricFamilies(m.url.String(), mfChan, m.transport); err != nil {
		handleError(w, NewHttpError("get metrics error", http.StatusInternalServerError))
		logger.Error().Err(err).Msg("get metrics error")
		return
	}

	result := []*prom2json.Family{}
	for mf := range mfChan {
		result = append(result, prom2json.NewFamily(mf))
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		handleError(w, NewHttpError("marshal json data error", http.StatusInternalServerError))
		logger.Error().Err(err).Msg("marshal json data error")
		return
	}

	w.Write(jsonData)
}

func handleError(w http.ResponseWriter, err *HttpError) {
	w.WriteHeader(err.Code())
	fmt.Fprintln(w, err.JsonSerialize())
}
