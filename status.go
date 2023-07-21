package echoserver

import (
	"github.com/aliworkshop/gateway"
	"net/http"
)

var statusMap = map[gateway.Status]int{
	gateway.StatusOK:        http.StatusOK,
	gateway.StatusCreated:   http.StatusCreated,
	gateway.StatusBadInput:  http.StatusBadRequest,
	gateway.StatusNoContent: http.StatusNoContent,
	gateway.StatusConflict:  http.StatusConflict,
}

func getStatusCode(status gateway.Status) int {
	if code, ok := statusMap[status]; ok {
		return code
	}
	return http.StatusNotImplemented
}
