package echoserver

import (
	"github.com/aliworkshop/handlerlib"
	"net/http"
)

var statusMap = map[handlerlib.Status]int{
	handlerlib.StatusOK:        http.StatusOK,
	handlerlib.StatusCreated:   http.StatusCreated,
	handlerlib.StatusBadInput:  http.StatusBadRequest,
	handlerlib.StatusNoContent: http.StatusNoContent,
	handlerlib.StatusConflict:  http.StatusConflict,
}

func getStatusCode(status handlerlib.Status) int {
	if code, ok := statusMap[status]; ok {
		return code
	}
	return http.StatusNotImplemented
}
