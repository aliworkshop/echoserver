package echoserver

import (
	"github.com/aliworkshop/error"
	"net/http"
)

func getStatusCodeByError(err error.ErrorModel) int {
	switch err.Type() {
	case error.TypeValidation:
		return http.StatusBadRequest
	case error.TypeNotFound:
		return http.StatusNotFound
	case error.TypeUnAuthorized:
		return http.StatusUnauthorized
	case error.TypeForbidden:
		return http.StatusForbidden
	case error.TypeTooManyRequests:
		return http.StatusTooManyRequests
	case error.TypeDuplicate:
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}
