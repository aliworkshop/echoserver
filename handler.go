package echoserver

import (
	"github.com/aliworkshop/handlerlib"
	"github.com/aliworkshop/loggerlib/logger"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type echoHandler struct {
	log             logger.Logger
	handlerFunction handlerlib.HandlerFunc
}

func NewHandler(logger logger.Logger) handlerlib.Handler {
	h := &echoHandler{
		log: logger,
	}
	return h
}

func (h *echoHandler) Upgrade(model handlerlib.RequestModel) (handlerlib.WebSocketModel, error) {
	return upgrade(model.GetContext().(echo.Context))
}
func (h *echoHandler) SetHandlerFunc(handler handlerlib.HandlerFunc) {
	h.handlerFunction = handler
}
func (h *echoHandler) HandlerFunc() handlerlib.HandlerFunc {
	return h.handlerFunction
}

func (h *echoHandler) Logger() logger.Logger {
	return h.log
}

func NewHandlerModel(logger logger.Logger, languageBundle *i18n.Bundle) (handlerModel handlerlib.HandlerModel) {
	baseHandler := NewHandler(logger)
	responder := NewResponder(languageBundle)
	return handlerlib.NewModel(baseHandler, responder)
}
