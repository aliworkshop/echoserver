package echoserver

import (
	"net/http"

	"github.com/aliworkshop/errors"
)

func getStatusCodeByError(err errors.ErrorModel) int {
	switch err.Type() {
	case errors.TypeValidation:
		return http.StatusBadRequest
	case errors.TypeNotFound:
		return http.StatusNotFound
	case errors.TypeUnAuthorized:
		return http.StatusUnauthorized
	case errors.TypeForbidden:
		return http.StatusForbidden
	case errors.TypeTooManyRequests:
		return http.StatusTooManyRequests
	case errors.TypeDuplicate:
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}
