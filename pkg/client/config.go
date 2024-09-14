package client

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultRetryAfter = time.Second * 10

type config struct {
	baseURL           *url.URL
	defaultRetryAfter time.Duration

	compress bool
	retry    bool

	transport http.RoundTripper
}

type ConfigOption func(*config)

func newConfig(address string, options ...ConfigOption) *config {
	config := &config{
		baseURL: &url.URL{
			Scheme: "http",
			Host:   address,
		},
		defaultRetryAfter: defaultRetryAfter,
		compress:          true,
		retry:             true,
		transport:         http.DefaultTransport,
	}

	if strings.Contains(address, "://") {
		url, err := url.Parse(address)
		if err == nil {
			config.baseURL = url
		}
	}

	for _, option := range options {
		option(config)
	}

	return config
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
