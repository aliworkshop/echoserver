package echoserver

import (
	"github.com/aliworkshop/error"
)

func getStatusCodeByError(err error.ErrorModel) int {
	switch err.Type() {
	case error.TypeValidation:
		return 400
	case error.TypeNotFound:
		return 404
	case error.TypeUnAuthorized:
		return 401
	case error.TypeForbidden:
		return 403
	case error.TypeTooManyRequests:
		return 429
	case error.TypeFailedDependency:
		return 424
	case error.TypeTooEarly:
		return 425
		//case error.TypeInternal:
		//	return ???
	}
	return 500
}
