package middleware

import (
	"fmt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"net/http"
)

func Recover() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				recovered := recover()

				switch recovered {
				case nil:
					return
				case http.ErrAbortHandler:
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(recovered)
				}

				err, ok := recovered.(error)
				if !ok {
					err = fmt.Errorf("%v", recovered)
				}

				controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Internal server error", err)
			}()

			next.ServeHTTP(writer, request)
		})
	}
}
