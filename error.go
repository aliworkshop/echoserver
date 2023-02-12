package echoserver

import (
	"github.com/aliworkshop/errorslib"
)

func getStatusCodeByError(err errorslib.ErrorModel) int {
	switch err.Type() {
	case errorslib.TypeValidation:
		return 400
	case errorslib.TypeNotFound:
		return 404
	case errorslib.TypeUnAuthorized:
		return 401
	case errorslib.TypeForbidden:
		return 403
	case errorslib.TypeTooManyRequests:
		return 429
	case errorslib.TypeFailedDependency:
		return 424
	case errorslib.TypeTooEarly:
		return 425
		//case errorslib.TypeInternal:
		//	return ???
	}
	return 500
}
