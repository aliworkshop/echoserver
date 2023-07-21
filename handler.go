package echoserver

import (
	"github.com/aliworkshop/gateway"
	"github.com/aliworkshop/loggerlib/logger"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type echoHandler struct {
	log             logger.Logger
	handlerFunction gateway.HandlerFunc
}

func NewHandler(logger logger.Logger) gateway.Handler {
	h := &echoHandler{
		log: logger,
	}
	return h
}

func (h *echoHandler) Upgrade(model gateway.Requester) (gateway.WebSocketHandler, error) {
	return upgrade(model.GetContext().(echo.Context))
}
func (h *echoHandler) SetHandlerFunc(handler gateway.HandlerFunc) {
	h.handlerFunction = handler
}
func (h *echoHandler) HandlerFunc() gateway.HandlerFunc {
	return h.handlerFunction
}

func (h *echoHandler) Logger() logger.Logger {
	return h.log
}

func NewHandlerEngine(logger logger.Logger, languageBundle *i18n.Bundle) (handlerModel gateway.HandlerEngine) {
	baseHandler := NewHandler(logger)
	responder := NewResponder(languageBundle)
	return gateway.NewEngine(baseHandler, responder)
}
