package client

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/http/retryafter"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/retry"
)

type Client struct {
	gzipPool *sync.Pool
	resty    *resty.Client
	config   *config
}

func New(address string, options ...ConfigOption) *Client {
	config := newConfig(address, options...)
	client := &Client{
		resty: resty.
			New().
			SetTransport(config.transport).
			SetBaseURL(config.baseURL.String()).
			SetHeader("Accept-Encoding", "gzip"),
		config: config,
	}

	if config.compress {
		client.gzipPool = &sync.Pool{
			New: func() any {
				writer, err := gzip.NewWriterLevel(io.Discard, 5)
				if err != nil {
					return nil
				}

				return writer
			},
		}
		client.resty.SetPreRequestHook(client.compressRequestBody)
	}

	return client
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
