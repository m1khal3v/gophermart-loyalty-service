package client

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
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

	transport http.RoundTripper
}

type ConfigOption func(*config)

func newConfig(address string, options ...ConfigOption) *config {
	config := &config{
		defaultRetryAfter: defaultRetryAfter,
		compress:          true,
		retry:             true,
		transport:         http.DefaultTransport,
	}

	if strings.Contains(address, "://") {
		url, err := url.Parse(address)
		if err != nil {
			panic(err)
		}

		host, port, err := net.SplitHostPort(url.Host)
		if err != nil {
			panic(err)
		}

		config.scheme = url.Scheme
		config.host = host
		config.port = port
	} else {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			config.host = address
		} else {
			config.scheme = "http"
			config.host = host
			config.port = port
		}
	}

	for _, option := range options {
		option(config)
	}

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
