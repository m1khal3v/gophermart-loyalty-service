package withdrawal

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	"net/http"
)

func (container *Container) List(writer http.ResponseWriter, request *http.Request) {
	userID, ok := context.UserIDFromContext(request.Context())
	if !ok {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t get request credentials", nil)
		return
	}

	has, err := container.manager.HasUser(request.Context(), userID)
	if err != nil {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t check that user has withdrawals", err)
		return
	}
	if !has {
		controller.WriteJSONResponse(http.StatusNoContent, responses.Message{
			Message: "withdrawals not found",
		}, writer)
		return
	}

	withdrawals, err := container.manager.FindByUser(request.Context(), userID)
	if err != nil {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t get user withdrawals", err)
		return
	}

	if err := controller.StreamJSONResponse(http.StatusOK, withdrawals, func(item *entity.Withdrawal) any {
		return responses.Withdrawal{
			Order:       item.OrderID,
			Sum:         item.Sum.AsFloat(),
			ProcessedAt: item.CreatedAt,
		}
	}, writer); err != nil {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t get user withdrawals", err)
		return
	}
}
