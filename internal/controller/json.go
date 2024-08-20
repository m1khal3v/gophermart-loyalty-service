package controller

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/responses"
	"go.uber.org/zap"
	"net/http"
)

func DecodeAndValidateJSONRequest[T any](request *http.Request, writer http.ResponseWriter) (*T, bool) {
	if request.Header.Get("Content-Type") != "application/json" {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid Content-Type", nil)
		return nil, false
	}

	target := new(T)

	if err := json.NewDecoder(request.Body).Decode(target); err != nil {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid json received", err)
		return nil, false
	}

	if _, err := govalidator.ValidateStruct(target); err != nil {
		WriteJSONErrorResponse(http.StatusBadRequest, writer, "Invalid request received", err)
		return nil, false
	}

	return target, true
}

func WriteJSONResponse(status int, response any, writer http.ResponseWriter) {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t encode response", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if _, err = writer.Write(jsonResponse); err != nil {
		WriteJSONErrorResponse(http.StatusInternalServerError, writer, "Can`t write response", err)
		return
	}
}

func StreamJSONResponse[T any](status int, stream <-chan T, transform func(item T) any, writer http.ResponseWriter) error {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if _, err := writer.Write([]byte("[")); err != nil {
		return err
	}

	first := true
	for item := range stream {
		if first {
			first = false
		} else {
			if _, err := writer.Write([]byte(",")); err != nil {
				return err
			}
		}

		jsonItem, err := json.Marshal(transform(item))
		if err != nil {
			return err
		}

		if _, err = writer.Write(jsonItem); err != nil {
			return err
		}
	}

	if _, err := writer.Write([]byte("]")); err != nil {
		return err
	}

	return nil
}

func WriteJSONErrorResponse(status int, writer http.ResponseWriter, message string, responseError error) {
	response := responses.APIError{
		Code:    status,
		Message: message,
	}

	if status >= 500 {
		logger.Logger.Error(message, zap.Error(responseError))
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logger.Logger.Error("Failed to encode response", zap.Error(err))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if _, err = writer.Write(jsonResponse); err != nil {
		logger.Logger.Error("Failed to write response", zap.Error(err))
	}
}
