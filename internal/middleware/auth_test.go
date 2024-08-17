package middleware

import (
	"fmt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateAuthorizationTokenOk(t *testing.T) {
	jwt := jwt.New("secret")
	userID := rand.Uint32N(1000) + 1
	token, err := jwt.Encode(userID, fmt.Sprintf("user_%d", userID))
	require.NoError(t, err)

	handler := ValidateAuthorizationToken(jwt)(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestUserID, ok := context.UserIDFromContext(request.Context())
		require.True(t, ok)
		assert.Equal(t, userID, requestUserID)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	writer := httptest.NewRecorder()

	handler.ServeHTTP(writer, request)
	assert.Equal(t, http.StatusOK, writer.Code)
}

func TestValidateAuthorizationTokenUnauthorized(t *testing.T) {
	jwt := jwt.New("secret")
	call := false

	handler := ValidateAuthorizationToken(jwt)(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		call = true
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer invalid")
	writer := httptest.NewRecorder()

	handler.ServeHTTP(writer, request)
	assert.Equal(t, http.StatusUnauthorized, writer.Code)
	assert.False(t, call)
}
