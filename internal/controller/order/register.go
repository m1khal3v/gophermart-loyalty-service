package order

import (
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/validator"
	"io"
	"net/http"
	"strconv"
)

func (container *Container) Register(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Content-Type") != "text/plain" {
		controller.WriteJSONErrorResponse(http.StatusBadRequest, writer, "invalid Content-Type", nil)
		return
	}

	userID, ok := context.UserIDFromContext(request.Context())
	if !ok {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t get request credentials", nil)
		return
	}

	id, err := io.ReadAll(request.Body)
	if err != nil {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t read request body", err)
		return
	}

	if !validator.IsLuhn(id) {
		controller.WriteJSONErrorResponse(http.StatusUnprocessableEntity, writer, "invalid order id", nil)
		return
	}

	uintID, err := strconv.ParseUint(string(id), 10, 64)
	if err != nil {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t parse id", err)
		return
	}

	order, err := container.orderManager.Register(request.Context(), uintID, userID)
	if err != nil {
		if errors.Is(err, manager.ErrOrderAlreadyRegisteredByCurrentUser) {
			controller.WriteJSONResponse(http.StatusOK, responses.Message{
				Message: "order already registered by current user",
			}, writer)
			return
		}
		if errors.Is(err, manager.ErrOrderAlreadyRegisteredByAnotherUser) {
			controller.WriteJSONErrorResponse(http.StatusConflict, writer, "order already registered by another user", err)
			return
		}

		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t register order", err)
		return
	}

	container.unprocessedQueue.Push(order.ID)
	controller.WriteJSONResponse(http.StatusAccepted, responses.Message{
		Message: "order has been successfully registered for processing",
	}, writer)
}
