package order

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

	has, err := container.orderManager.HasUser(request.Context(), userID)
	if err != nil {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t check that user has orders", err)
		return
	}
	if !has {
		controller.WriteJSONResponse(http.StatusNoContent, responses.Message{
			Message: "orders not found",
		}, writer)
		return
	}

	orders, err := container.orderManager.FindByUser(request.Context(), userID)
	if err != nil {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t get user orders", err)
		return
	}

	if err := controller.StreamJSONResponse(http.StatusOK, orders, func(item *entity.Order) any {
		response := responses.Order{
			Number:     item.ID,
			Status:     item.Status,
			UploadedAt: item.CreatedAt,
		}
		if item.Status == entity.OrderStatusProcessed {
			accrual := item.Accrual.AsFloat()
			response.Accrual = &accrual
		}

		return response
	}, writer); err != nil {
		controller.WriteJSONErrorResponse(http.StatusInternalServerError, writer, "can`t get user orders", err)
		return
	}
}
