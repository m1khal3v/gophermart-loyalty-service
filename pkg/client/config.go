package client

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const defaultRetryAfter = time.Second * 10

type config struct {
	scheme            string
	host              string
	port              string
	defaultRetryAfter time.Duration

	compress bool
	retry    bool

	address   string
	transport http.RoundTripper
}

type ConfigOption func(*config)

func newConfig(address string, options ...ConfigOption) *config {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = "80"
	}

	config := &config{
		scheme:            "http",
		host:              host,
		port:              port,
		defaultRetryAfter: defaultRetryAfter,
		compress:          true,
		retry:             true,
		transport:         http.DefaultTransport,
	}

	for _, option := range options {
		option(config)
	}

	config.address = fmt.Sprintf("%s://%s", config.scheme, strings.TrimRight(net.JoinHostPort(config.host, config.port), ":"))

	return config
}

func WithScheme(scheme string) ConfigOption {
	return func(config *config) {
		config.scheme = scheme
	}
}

func WithPort(port uint32) ConfigOption {
	return func(config *config) {
		config.port = fmt.Sprintf("%d", port)
	}
}

func WithoutCompress() ConfigOption {
	return func(config *config) {
		config.compress = false
	}
}

func WithoutRetry() ConfigOption {
	return func(config *config) {
		config.retry = false
	}
}

func WithDefaultRetryAfter(retryAfter time.Duration) ConfigOption {
	return func(config *config) {
		config.defaultRetryAfter = retryAfter
	}
}

func withTransport(transport http.RoundTripper) ConfigOption {
	return func(config *config) {
		config.transport = transport
	}
}
