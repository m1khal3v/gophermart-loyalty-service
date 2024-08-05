package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/retry"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// This is a very simplified regular expression that will work in most cases.
// In border cases, you can disable address verification through the config
var addressRegex = regexp.MustCompile(`^https?://[a-zA-Z0-9][a-zA-Z0-9-.]*(:\d+)?(/[a-zA-Z0-9-_+%]*)*$`)

type Config struct {
	Address string

	DisableCompress          bool
	DisableAddressValidation bool
	DisableRetry             bool
}

type Client struct {
	resty  *resty.Client
	config *Config
}

type ErrUnexpectedStatus struct {
	Status int
}

func (err ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status code: %d", err.Status)
}

func newErrUnexpectedStatus(status int) ErrUnexpectedStatus {
	return ErrUnexpectedStatus{
		Status: status,
	}
}

type ErrInvalidAddress struct {
	Address string
}

func (err ErrInvalidAddress) Error() string {
	return fmt.Sprintf("invalid address: %s", err.Address)
}

func newErrInvalidAddress(address string) ErrInvalidAddress {
	return ErrInvalidAddress{
		Address: address,
	}
}

func New(config *Config) (*Client, error) {
	if err := prepareConfig(config); err != nil {
		return nil, err
	}

	client := resty.
		New().
		SetBaseURL(config.Address).
		SetHeader("Accept-Encoding", "gzip")

	if !config.DisableCompress {
		client.SetPreRequestHook(compressRequestBody)
	}

	return &Client{resty: client, config: config}, nil
}

func prepareConfig(config *Config) error {
	if !config.DisableAddressValidation {
		if !strings.HasPrefix(config.Address, "http") {
			config.Address = "http://" + config.Address
		}

		if !addressRegex.MatchString(config.Address) {
			return newErrInvalidAddress(config.Address)
		}
	}

	return nil
}

func compressRequestBody(client *resty.Client, request *http.Request) error {
	if request.Body == nil {
		return nil
	}

	buffer := bytes.NewBuffer([]byte{})
	writer, err := gzip.NewWriterLevel(buffer, 5)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, request.Body)
	if err = errors.Join(err, writer.Close(), request.Body.Close()); err != nil {
		return err
	}

	request.Body = io.NopCloser(buffer)
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}
	request.ContentLength = int64(buffer.Len())
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))

	return nil
}

func (client *Client) createRequest(ctx context.Context) *resty.Request {
	return client.resty.R().SetContext(ctx)
}

func (client *Client) doRequest(request *resty.Request, method, url string) (*resty.Response, error) {
	var result *resty.Response = nil
	do := func() error {
		var err error
		result, err = request.Execute(method, url)
		if err != nil {
			return err
		}

		if result.StatusCode() != http.StatusOK {
			return newErrUnexpectedStatus(result.StatusCode())
		}

		return nil
	}

	var err error
	if !client.config.DisableRetry {
		err = retry.Retry(time.Second, 5*time.Second, 4, 2, do, func(err error) bool {
			return !errors.As(err, &ErrUnexpectedStatus{}) &&
				!errors.Is(err, context.DeadlineExceeded) &&
				!errors.Is(err, context.Canceled)
		})
	} else {
		err = do()
	}

	return result, err
}
