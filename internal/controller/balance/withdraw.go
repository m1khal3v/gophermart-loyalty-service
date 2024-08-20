package balance

import (
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/requests"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	"net/http"
)

func (container *Container) Withdraw(writer http.ResponseWriter, request *http.Request) {
	userID, ok := context.UserIDFromContext(request.Context())
	if !ok {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t get request credentials", nil)
		return
	}

	withdrawRequest, ok := controller.DecodeAndValidateJSONRequest[requests.Withdraw](request, writer)
	if !ok {
		return
	}

	if err := container.userWithdrawalManager.Withdraw(request.Context(), withdrawRequest.Order, userID, withdrawRequest.Sum); err != nil {
		if errors.Is(err, manager.ErrInsufficientFunds) {
			controller.WriteJSONErrorResponse(http.StatusPaymentRequired, writer, "insufficient funds", err)
		} else {
			controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t register withdrawal", err)
		}

		return
	}

	controller.WriteJSONResponse(http.StatusOK, responses.Message{
		Message: "withdrawal successfully registered",
	}, writer)
}
