package middleware

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/controller"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"net/http"
	"strings"
)

func ValidateAuthorizationToken(jwt *jwt.Container) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			token := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
			claims, err := jwt.Decode(token)
			if err != nil {
				controller.WriteJSONErrorResponse(http.StatusUnauthorized, writer, "invalid credentials", err)
				return
			}

			request = request.WithContext(context.WithUserID(request.Context(), claims.SubjectID))

			next.ServeHTTP(writer, request)
		})
	}
}
