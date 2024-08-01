package user

import (
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/requests"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	"net/http"
)

func (container *Container) Register(writer http.ResponseWriter, request *http.Request) {
	registerRequest, ok := controller.DecodeAndValidateJSONRequest[requests.RegisterUser](request, writer)
	if !ok {
		return
	}

	user, err := container.manager.RegisterUser(registerRequest.Login, registerRequest.Password)
	if err != nil {
		if errors.Is(err, manager.ErrLoginAlreadyExists) {
			controller.WriteJSONErrorResponse(http.StatusConflict, writer, "login already exists", err)
		} else {
			controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "internal server error", err)
		}

		return
	}

	controller.WriteJSONResponse(responses.RegisterUser{
		ID:    user.ID,
		Login: user.Login,
	}, writer)
}
