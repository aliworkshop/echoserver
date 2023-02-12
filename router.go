package echoserver

import (
	"context"
	"fmt"
	"github.com/aliworkshop/handlerlib"
	"github.com/labstack/echo/v4"
	"strings"
)

type router struct {
	config config
}

func (rh *router) getContext(request handlerlib.RequestModel) (context.Context, context.CancelFunc) {
	ctx := request.GetConnectionContext()
	// IP
	ip := request.GetHeader("X-Forwarded-For")
	ip = strings.TrimSpace(strings.Split(ip, ",")[0])
	if ip == "" {
		ip = strings.TrimSpace(request.GetHeader("X-Real-Ip"))
	}
	if ip == "" {
		ip = strings.TrimSpace(strings.Split(request.BaseRequest().RemoteAddr, ":")[0])
	}
	ctx = context.WithValue(ctx, "IP", ip)

	return context.WithTimeout(ctx, rh.config.ConnectionTimeout)
}

func (rh *router) getHandler(monitoring handlerlib.MonitoringModel, handler handlerlib.HandlerModel,
	isFirstHandler, isLastHandler, shouldRespond bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req handlerlib.RequestModel

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
			l := handler.Logger()
			req = NewRequest(c, handler.LanguageBundle()).WithLogger(l)
			ctx, cancel := rh.getContext(req)
			req.SetConnectionContext(ctx)
			c.Set("req", req)
			c.Set("cancel", cancel)
			// monitoring beginning of request
			monitoring.OnRequestStart(req)
		} else {
			// load from context
			_req := c.Get("req")
			if q, ok := _req.(handlerlib.RequestModel); ok {
				req = q
			} else {
				// mark as last handler
				isLastHandler = true
				return nil
			}
		}
		session := req.GetSession()
		if session != nil {
			if err := rh.verifySession(c, session); err != nil {
				isLastHandler = true
				return echo.ErrForbidden
			}
		}
		// respond only on last handler, and if is not responded yet
		handlerlib.Handle(handler, req, shouldRespond)
		return nil
	}
}

func (rh *router) getMiddleware(monitoring handlerlib.MonitoringModel, handler handlerlib.HandlerModel,
	isFirstHandler, isLastHandler bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var req handlerlib.RequestModel

			if isFirstHandler {
				// only create request on first handler function call
				l := handler.Logger()
				req = NewRequest(c, handler.LanguageBundle()).WithLogger(l)
				ctx, cancel := rh.getContext(req)
				req.SetConnectionContext(ctx)
				c.Set("req", req)
				c.Set("cancel", cancel)
				// monitoring beginning of request
				monitoring.OnRequestStart(req)
			} else {
				// load from context
				_req := c.Get("req")
				if q, ok := _req.(handlerlib.RequestModel); ok {
					req = q
				} else {
					// mark as last handler
					isLastHandler = true
					return nil
				}
			}
			session := req.GetSession()
			if session != nil {
				if err := rh.verifySession(c, session); err != nil {
					isLastHandler = true
					return echo.ErrForbidden
				}
			}
			if handlerlib.Handle(handler, req, false) {
				return next(c)
			}
			return nil
		}
	}
}

func (rh *router) match(monitoring handlerlib.MonitoringModel,
	hs ...handlerlib.HandlerModel) (echo.HandlerFunc, []echo.MiddlewareFunc) {
	var hf echo.HandlerFunc
	var mfs []echo.MiddlewareFunc
	if len(hs) == 1 {
		hf = rh.getHandler(monitoring, hs[0], true, true, true)
		return hf, mfs
	}
	for i, h := range hs[:len(hs)-1] {
		if h.Logger() == nil {
			fmt.Println(fmt.Sprintf("logger is not set for "+
				"handler: %+v, Call to log panics\n", h))
		}
		isLastHandler := i == len(hs)-1
		mf := rh.getMiddleware(monitoring, h, i == 0, isLastHandler)
		mfs = append(mfs, mf)
	}
	hf = rh.getHandler(monitoring, hs[len(hs)-1], false, true, true)

	return hf, mfs
}
