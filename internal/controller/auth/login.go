package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/requests"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
)

func (container *Container) Login(writer http.ResponseWriter, request *http.Request) {
	registerRequest, ok := controller.DecodeAndValidateJSONRequest[requests.Login](request, writer)
	if !ok {
		return
	}

	token, err := container.manager.Authorize(request.Context(), registerRequest.Login, registerRequest.Password)
	if err != nil {
		if errors.Is(err, manager.ErrInvalidCredentials) {
			controller.WriteJSONErrorResponse(http.StatusUnauthorized, writer, "invalid credentials", err)
		} else {
			controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "internal server error", err)
		}

		return
	}

	writer.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token)) // for autotests
	controller.WriteJSONResponse(http.StatusOK, responses.Auth{
		AccessToken: token,
	}, writer)
}
