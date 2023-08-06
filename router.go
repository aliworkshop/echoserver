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

func (rh *router) getHandler(monitoring gateway.MonitoringModel, controller gateway.Controller, handler gateway.Handler,
	isFirstHandler, isLastHandler, shouldRespond bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req gateway.Requester

		defer func() {
			if isLastHandler {
				cancel := c.Get("cancel")
				if cancel != nil {
					defer cancel.(context.CancelFunc)()
				}
				monitoring.OnRequestEnd(req)
			}
		}()

		if isFirstHandler {
			// only create request on first handler function call
			req = NewRequest(c, controller.LanguageBundle())
			ctx, cancel := rh.getContext(req)
			req.SetConnectionContext(ctx)
			c.Set("req", req)
			c.Set("cancel", cancel)
			// monitoring beginning of request
			monitoring.OnRequestStart(req)
		} else {
			// load from context
			_req := c.Get("req")
			if q, ok := _req.(gateway.Requester); ok {
				req = q
			} else {
				// mark as last handler
				isLastHandler = true
				return nil
			}
		}

		controller.Process(handler, req, shouldRespond)
		// respond only on last handler, and if is not responded yet
		return nil
	}
}

func (rh *router) getMiddleware(monitoring gateway.MonitoringModel, controller gateway.Controller, handler gateway.Handler,
	isFirstHandler, isLastHandler bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var req gateway.Requester

			if isFirstHandler {
				// only create request on first handler function call
				req = NewRequest(c, controller.LanguageBundle())
				ctx, cancel := rh.getContext(req)
				req.SetConnectionContext(ctx)
				c.Set("req", req)
				c.Set("cancel", cancel)
				// monitoring beginning of request
				monitoring.OnRequestStart(req)
			} else {
				// load from context
				_req := c.Get("req")
				if q, ok := _req.(gateway.Requester); ok {
					req = q
				} else {
					// mark as last handler
					isLastHandler = true
					return nil
				}
			}

			if controller.Process(handler, req, false) {
				return next(c)
			}
			return nil
		}
	}
}

func (rh *router) match(monitoring gateway.MonitoringModel, c gateway.Controller,
	hs ...gateway.Handler) (echo.HandlerFunc, []echo.MiddlewareFunc) {
	var hf echo.HandlerFunc
	var mfs []echo.MiddlewareFunc
	if len(hs) == 1 {
		hf = rh.getHandler(monitoring, c, hs[0], true, true, true)
		return hf, mfs
	}
	for i, h := range hs[:len(hs)-1] {
		isLastHandler := i == len(hs)-1
		mf := rh.getMiddleware(monitoring, c, h, i == 0, isLastHandler)
		mfs = append(mfs, mf)
	}
	hf = rh.getHandler(monitoring, c, hs[len(hs)-1], false, true, true)

	return hf, mfs
}

func (rh *router) matchMiddleware(monitoring gateway.MonitoringModel, c gateway.Controller,
	hs ...gateway.Handler) []echo.MiddlewareFunc {
	var mfs []echo.MiddlewareFunc
	for i, h := range hs {
		mf := rh.getMiddleware(monitoring, c, h, i == 0, i == len(hs)-1)
		mfs = append(mfs, mf)
	}

	return mfs
}
