package echoserver

import (
	"errors"

	"github.com/aliworkshop/gateway/v2"
	"github.com/labstack/echo/v4"
)

type router struct {
	config config
}

func (rh *router) getHandler(controller gateway.Controller, handler gateway.Handler, shouldRespond bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		if handler == nil {
			return errors.New("no handler is defined for this route")
		}
		req := getOrCreateRequest(c, controller)
		controller.Process(handler, req, shouldRespond)
		return nil
	}
}

func (rh *router) getMiddleware(controller gateway.Controller, handler gateway.Handler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := getOrCreateRequest(c, controller)
			if controller.Process(handler, req, false) {
				return next(c)
			}
			return nil
		}
	}
}

func (rh *router) match(c gateway.Controller, hs ...gateway.Handler) (echo.HandlerFunc, []echo.MiddlewareFunc) {
	var hf echo.HandlerFunc
	var mfs []echo.MiddlewareFunc
	if len(hs) == 1 {
		hf = rh.getHandler(c, hs[0], true)
		return hf, mfs
	}
	for _, h := range hs[:len(hs)-1] {
		mfs = append(mfs, rh.getMiddleware(c, h))
	}
	hf = rh.getHandler(c, hs[len(hs)-1], true)
	return hf, mfs
}

func (rh *router) matchMiddleware(c gateway.Controller, hs ...gateway.Handler) []echo.MiddlewareFunc {
	mfs := make([]echo.MiddlewareFunc, 0, len(hs))
	for _, h := range hs {
		mfs = append(mfs, rh.getMiddleware(c, h))
	}
	return mfs
}

func getOrCreateRequest(c echo.Context, controller gateway.Controller) gateway.HttpRequester {
	if v := c.Get("req"); v != nil {
		if req, ok := v.(gateway.HttpRequester); ok {
			return req
		}
	}
	req := NewRequest(c, controller.LanguageBundle())
	c.Set("req", req)
	return req
}
