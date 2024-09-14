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
		},
		defaultRetryAfter: defaultRetryAfter,
		compress:          true,
		retry:             true,
		transport:         http.DefaultTransport,
	}

	resolveAddress(address, config)

	for _, option := range options {
		option(config)
	}

	return config
}

func resolveAddress(address string, config *config) {
	if strings.Contains(address, "://") {
		url, err := url.Parse(address)
		if err != nil {
			panic(err)
		}

		config.baseURL = url
	} else {
		config.baseURL.Host = address
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
