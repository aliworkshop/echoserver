package echoserver

import (
	"context"
	"github.com/aliworkshop/gateway/v2"
	"github.com/labstack/echo/v4"
	"strings"
)

type router struct {
	config config
}

func (rh *router) getContext(request gateway.Requester) (context.Context, context.CancelFunc) {
	ctx := request.GetConnectionContext()
	// IP
	ip := request.GetHeader("X-Forwarded-For")
	ip = strings.TrimSpace(strings.Split(ip, ",")[0])
	if ip == "" {
		ip = strings.TrimSpace(request.GetHeader("X-Real-Ip"))
	}
	if ip == "" {
		ip = strings.TrimSpace(strings.Split(request.Request().RemoteAddr, ":")[0])
	}
	ctx = context.WithValue(ctx, "IP", ip)

	return context.WithTimeout(ctx, rh.config.ConnectionTimeout)
}

func (rh *router) getHandler(controller gateway.Controller, handler gateway.Handler, shouldRespond bool) echo.HandlerFunc {
	return func(c echo.Context) error {

		var req gateway.Requester
		if _req := c.Get("req"); _req != nil {
			req = _req.(gateway.Requester)
		} else {
			req = NewRequest(c, controller.LanguageBundle())
			c.Set("req", req)
		}
		controller.Process(handler, req, shouldRespond)
		return nil
	}
}

func (rh *router) getMiddleware(controller gateway.Controller, handler gateway.Handler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			var req gateway.Requester
			if _req := c.Get("req"); _req != nil {
				req = _req.(gateway.Requester)
			} else {
				req = NewRequest(c, controller.LanguageBundle())
				c.Set("req", req)
			}
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
		mf := rh.getMiddleware(c, h)
		mfs = append(mfs, mf)
	}
	hf = rh.getHandler(c, hs[len(hs)-1], true)

	return hf, mfs
}

func (rh *router) matchMiddleware(c gateway.Controller, hs ...gateway.Handler) []echo.MiddlewareFunc {
	var mfs []echo.MiddlewareFunc
	for _, h := range hs {
		mf := rh.getMiddleware(c, h)
		mfs = append(mfs, mf)
	}

	return mfs
}
