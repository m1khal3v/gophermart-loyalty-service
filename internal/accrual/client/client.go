package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/http/retryafter"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/retry"
)

type Client struct {
	resty  *resty.Client
	config *config
}

func New(address string, options ...ConfigOption) *Client {
	config := newConfig(address, options...)

	client := resty.
		New().
		SetTransport(config.transport).
		SetScheme(config.scheme).
		SetBaseURL(strings.TrimRight(net.JoinHostPort(config.host, config.port), ":")).
		SetHeader("Accept-Encoding", "gzip")

	if config.compress {
		client.SetPreRequestHook(compressRequestBody)
	}

	return &Client{resty: client, config: config}
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
			return newErrTooManyRequests(retryafter.Parse(result.Header().Get("Retry-After"), client.config.defaultRetryAfter))
		case http.StatusInternalServerError:
			return ErrInternalServerError
		default:
			return newErrUnexpectedStatus(result.StatusCode())
		}
	}

	var err error
	if client.config.retry {
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
