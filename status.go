package echoserver

import (
	"github.com/aliworkshop/gateway/v2"
	"net/http"
)

var statusMap = map[gateway.Status]int{
	gateway.StatusOK:                http.StatusOK,
	gateway.StatusCreated:           http.StatusCreated,
	gateway.StatusNoContent:         http.StatusNoContent,
	gateway.StatusMovedPermanently:  http.StatusMovedPermanently,
	gateway.StatusFound:             http.StatusFound,
	gateway.StatusPermanentRedirect: http.StatusPermanentRedirect,
	gateway.StatusTemporaryRedirect: http.StatusTemporaryRedirect,
	gateway.StatusBadInput:          http.StatusBadRequest,
	gateway.StatusConflict:          http.StatusConflict,
}

func getStatusCode(status gateway.Status) int {
	if code, ok := statusMap[status]; ok {
		return code
	}
	return http.StatusNotImplemented
}
