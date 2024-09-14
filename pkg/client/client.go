package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/http/retryafter"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/requests"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
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
		SetBaseURL(config.address).
		SetHeader("Accept-Encoding", "gzip")

	if config.compress {
		client.SetPreRequestHook(compressRequestBody)
	}

	return &Client{resty: client, config: config}
}

func (client *Client) Register(ctx context.Context, request *requests.Register) (*responses.Auth, *responses.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		SetResult(&responses.Auth{}),
		resty.MethodPost, "api/user/register")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*responses.APIError), err
		}
	}

	return result.Result().(*responses.Auth), nil, nil
}

func (client *Client) Login(ctx context.Context, request *requests.Login) (*responses.Auth, *responses.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		SetResult(&responses.Auth{}),
		resty.MethodPost, "api/user/login")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*responses.APIError), err
		}
	}

	return result.Result().(*responses.Auth), nil, nil
}

func (client *Client) AddOrder(ctx context.Context, token string, orderID uint64) (*responses.Message, *responses.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "text/plain").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		SetBody(strconv.FormatUint(orderID, 10)).
		SetResult(&responses.Message{}),
		resty.MethodPost, "api/user/orders")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*responses.APIError), err
		}
	}

	return result.Result().(*responses.Message), nil, nil
}

func (client *Client) Orders(ctx context.Context, token string) ([]responses.Order, *responses.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		SetResult(&[]responses.Order{}),
		resty.MethodGet, "api/user/orders")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*responses.APIError), err
		}
	}

	return *result.Result().(*[]responses.Order), nil, nil
}

func (client *Client) Withdrawals(ctx context.Context, token string) ([]responses.Withdrawal, *responses.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		SetResult(&[]responses.Withdrawal{}),
		resty.MethodGet, "api/user/withdrawals")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*responses.APIError), err
		}
	}

	return *result.Result().(*[]responses.Withdrawal), nil, nil
}

func (client *Client) Balance(ctx context.Context, token string) (*responses.Balance, *responses.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		SetResult(&responses.Balance{}),
		resty.MethodGet, "api/user/balance")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*responses.APIError), err
		}
	}

	return result.Result().(*responses.Balance), nil, nil
}

func (client *Client) Withdraw(ctx context.Context, token string, request *requests.Withdraw) (*responses.Message, *responses.APIError, error) {
	result, err := client.doRequest(client.createRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
		SetBody(request).
		SetResult(&responses.Message{}),
		resty.MethodPost, "api/user/balance/withdraw")

	if err != nil {
		if result == nil || result.RawResponse == nil {
			return nil, nil, err
		} else {
			return nil, result.Error().(*responses.APIError), err
		}
	}

	return result.Result().(*responses.Message), nil, nil
}

func (client *Client) createRequest(ctx context.Context) *resty.Request {
	return client.resty.R().SetContext(ctx).SetError(&responses.APIError{})
}

func (client *Client) doRequest(request *resty.Request, method, url string) (*resty.Response, error) {
	var result *resty.Response
	do := func() error {
		var err error
		result, err = request.Execute(method, url)
		if err != nil {
			return err
		}

		switch status := result.StatusCode(); {
		case status >= http.StatusOK && status < http.StatusMultipleChoices:
			return nil
		case status == http.StatusBadRequest:
			return ErrBadRequest
		case status == http.StatusUnauthorized:
			return ErrInvalidCredentials
		case status == http.StatusTooManyRequests:
			return newErrTooManyRequests(retryafter.Parse(result.Header().Get("Retry-After"), client.config.defaultRetryAfter))
		case status == http.StatusInternalServerError:
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
				!errors.Is(err, ErrInvalidCredentials) &&
				!errors.Is(err, context.DeadlineExceeded) &&
				!errors.Is(err, context.Canceled)
		})
	} else {
		err = do()
	}

	return result, err
}
