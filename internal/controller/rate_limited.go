package controller

import (
	"errors"
	"net/http"
)

var ErrTooManyRequests = errors.New("too many requests")

func RateLimited(writer http.ResponseWriter, request *http.Request) {
	WriteJSONErrorResponse(http.StatusTooManyRequests, writer, "too many requests", ErrTooManyRequests)
}
