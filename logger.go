package echoserver

import (
	"github.com/aliworkshop/handlerlib"
	"github.com/aliworkshop/loggerlib/logger"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

func NewLoggerHandler(l logger.Logger, serverConfig Http) echo.MiddlewareFunc {
	skipPaths := make(map[string]string, 0)
	for _, sp := range serverConfig.Logger.SkipPaths {
		skipPaths[sp] = ""
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			_, ok := skipPaths[c.Path()]
			if !ok || c.Path() == "" {
				path := c.Request().URL.Path
				raw := c.Request().URL.RawQuery
				if raw != "" {
					path = path + "?" + raw
				}
				status := c.Response().Status

				uid := c.Get("UID")
				if uid != nil {
					l = l.WithUid(uid.(string))
				}

				l = l.WithSource(serverConfig.ServiceName).WithId("echoServer")
				meta := logger.Field{
					"Path":       path,
					"Ip":         c.RealIP(),
					"Elapsed":    time.Since(start).Milliseconds(),
					"Method":     c.Request().Method,
					"StatusCode": status,
					"Mode":       c.Request().Header.Values("X-Mode"),
				}
				if req := c.Get("req"); req != nil {
					request := req.(handlerlib.RequestModel)
					if userId := request.GetCurrentAccountId(); userId != nil {
						meta["UserId"] = userId
					}
				}
				if status < 400 {
					l.With(meta).DebugF("")
				} else if status < 500 {
					if status == http.StatusFailedDependency {
						l.With(meta).CriticalF("StatusFailedDependency")
					} else {
						l.With(meta).InfoF("")
					}
				} else {
					l.With(meta).CriticalF("")
				}
			}
			return next(c)
		}
	}
}
