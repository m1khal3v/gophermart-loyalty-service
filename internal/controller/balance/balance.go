package balance

import (
	"net/http"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
)

func (container *Container) Balance(writer http.ResponseWriter, request *http.Request) {
	userID, ok := context.UserIDFromContext(request.Context())
	if !ok {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t get request credentials", nil)
		return
	}

	user, err := container.userManager.FindByID(request.Context(), userID)
	if err != nil {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t get user", err)
		return
	}

	controller.WriteJSONResponse(http.StatusOK, responses.Balance{
		Current:   user.Balance.AsFloat(),
		Withdrawn: user.Withdrawn.AsFloat(),
	}, writer)
}
