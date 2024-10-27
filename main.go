package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/prom2json"
)

type Config struct {
	// server config
	ServerPort int `json:"server_port"`
	// metrics config
	MetricsURLStr string      `json:"metrics_url"`
	MetricsURL    *url.URL    `json:"-"`
	TLSConfig     *tls.Config `json:"tls_config"`
}

const (
	ExitCodeOK     int = 0
	ExitCodeError  int = 1
	ExitCodeMisuse int = 2
)

const (
	ForceKillServerTimeout = 5 * time.Second
)

type Error interface {
	Error() string
	ExitCode() int
	Err() error
}

type SError struct {
	err      error
	exitCode int
}

func (e *SError) Error() string {
	return e.err.Error()
}

func (e *SError) ExitCode() int {
	return e.exitCode
}

func (e *SError) Err() error {
	return e.err
}

func NewError(code int, err error) Error {
	return &SError{
		err:      err,
		exitCode: code,
	}
}

func NewErrorf(code int, format string, args ...interface{}) Error {
	return &SError{
		err:      fmt.Errorf(format, args...),
		exitCode: code,
	}
}

func main() {
	cfg, err := ParseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(err.ExitCode())
	}

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	serverErr := startServer(ctx, cfg)
	done()

	if serverErr != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(err.ExitCode())
	}
}

var usage = fmt.Sprintf(`Usage: %s [METRICS_PATH | METRICS_URL [--cert CERT_PATH --key KEY_PATH | --accept-invalid-cert]]

Example:

	$ prom2json http://my-prometheus-server:9000/metrics

	$ curl http://my-prometheus-server:9000/metrics | prom2json
	
`, os.Args[0])

func ParseConfig() (*Config, Error) {
	serverPort := flag.Int("port", 8080, "Server port")
	tlsCert := flag.String("tls-cert", "", "client TLS certificate file")
	tlsKey := flag.String("tls-key", "", "client TLS certificate's key file")
	TLSSkipVerifyCert := flag.Bool("tls-skip-verify-cert", false, "Accept any certificate during TLS handshake. Insecure, use only for testing.")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	metricsUrl := flag.Arg(0)
	flag.NArg()

	if flag.NArg() > 1 {
		return nil, NewErrorf(ExitCodeMisuse, "Too many arguments.\n%s", usage)
	}

	if metricsUrl == "" {
		return nil, NewErrorf(ExitCodeMisuse, "Missing METRICS_URL argument.\n%s", usage)
	}

	url, urlErr := url.Parse(metricsUrl)
	if urlErr != nil || url.Scheme == "" {
		return nil, NewErrorf(ExitCodeMisuse, "Invalid METRICS_URL argument.\n%s", usage)
	}

	// validate TLS arguments
	if (*tlsCert != "" && *tlsKey == "") || (*tlsCert == "" && *tlsKey != "") {
		return nil, NewErrorf(ExitCodeError, "Missing TLS certificate or key.\n%s", usage)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: *TLSSkipVerifyCert,
	}
	if *tlsCert != "" && *tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			return nil, NewError(ExitCodeError, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return &Config{
		ServerPort:    *serverPort,
		MetricsURLStr: metricsUrl,
		MetricsURL:    url,
		TLSConfig:     tlsConfig,
	}, nil
}

func startServer(ctx context.Context, cfg *Config) Error {
	http.HandleFunc("/metrics-json", metricsHandler)
	svr := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.ServerPort),
	}

	// Start the HTTP server on port 8080
	log.Printf("Starting server on port [%d]", cfg.ServerPort)

	// Handle grateful shutdown
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()

		log.Println("Context closed, shutting down server")
		shutdownCtx, done := context.WithTimeout(context.Background(), ForceKillServerTimeout)
		defer done()
		errCh <- svr.Shutdown(shutdownCtx)
	}()

	// If server serve failed with error other than http.ErrServerClosed, return the error
	// Otherwise, wait for the server to shutdown gracefully
	if err := svr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return NewError(ExitCodeError, err)
	}

	log.Println("Server stopped serving, waiting for graceful shutdown")

	// Wait for the server to shutdown gracefully & return error if any
	if err := <-errCh; err != nil {
		return NewError(ExitCodeError, err)
	}
	return nil
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// URL of the Prometheus metrics endpoint
	prometheusURL := "http://your-prometheus-agent.example.org:9090/metrics"

	// Fetch metrics from the Prometheus agent
	resp, err := http.Get(prometheusURL)
	if err != nil {
		http.Error(w, "Error fetching metrics from Prometheus", http.StatusInternalServerError)
		log.Printf("Error fetching metrics: %v", err)
		return
	}
	defer resp.Body.Close()

	// Parse the Prometheus metrics to JSON
	metrics := prom2json.ParseResponse(resp.Body)

	// Convert the parsed metrics to JSON format
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		http.Error(w, "Error converting metrics to JSON", http.StatusInternalServerError)
		log.Printf("Error converting metrics to JSON: %v", err)
		return
	}

	// Set the content type to application/json and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
