package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/http/retryafter"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/retry"
	"io"
	"net/http"
	"regexp"
	"strconv"
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

func (client *Client) GetAccrual(ctx context.Context, orderID uint64) (*responses.Accrual, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetPathParams(map[string]string{
			"orderID": strconv.FormatUint(orderID, 10),
		}).
		SetResult(&responses.Accrual{}),
		resty.MethodGet, "api/orders/{orderID}")

	if err != nil {
		return nil, err
	}

	return result.Result().(*responses.Accrual), nil
}

func (client *Client) createRequest(ctx context.Context) *resty.Request {
	return client.resty.R().SetContext(ctx)
}

func (client *Client) doRequest(request *resty.Request, method, url string) (*resty.Response, error) {
	var result *resty.Response
	do := func() error {
		var err error
		result, err = request.Execute(method, url)
		if err != nil {
			return err
		}

		switch result.StatusCode() {
		case http.StatusOK:
			return nil
		case http.StatusNoContent:
			return ErrOrderNotFound
		case http.StatusTooManyRequests:
			return newErrTooManyRequests(retryafter.Parse(result.Header().Get("Retry-After"), time.Second*10))
		case http.StatusInternalServerError:
			return ErrInternalServerError
		default:
			return newErrUnexpectedStatus(result.StatusCode())
		}
	}

	var err error
	if !client.config.DisableRetry {
		err = retry.Retry(time.Second, 5*time.Second, 4, 2, do, func(err error) bool {
			return !errors.As(err, &ErrUnexpectedStatus{}) &&
				!errors.As(err, &ErrTooManyRequests{}) &&
				!errors.Is(err, context.DeadlineExceeded) &&
				!errors.Is(err, context.Canceled)
		})
	} else {
		err = do()
	}

	return result, err
}
