package echoserver

import (
	"net/http"
	"time"

	"github.com/aliworkshop/gateway/v2"
	"github.com/aliworkshop/logger"
	"github.com/labstack/echo/v4"
)

func NewLoggerHandler(l logger.Logger, serverConfig Http) echo.MiddlewareFunc {
	skipPaths := make(map[string]struct{}, len(serverConfig.Logger.SkipPaths))
	for _, sp := range serverConfig.Logger.SkipPaths {
		skipPaths[sp] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			if err := next(c); err != nil {
				c.Error(err)
			}
			if _, skip := skipPaths[c.Path()]; skip && c.Path() != "" {
				return nil
			}
			path := c.Request().URL.Path
			if raw := c.Request().URL.RawQuery; raw != "" {
				path = path + "?" + raw
			}
			status := c.Response().Status

			line := l
			if uid := c.Get("UID"); uid != nil {
				line = line.WithUid(uid.(string))
			}
			line = line.WithSource(serverConfig.ServiceName).WithId("echoServer")

			meta := logger.Field{
				"Path":       path,
				"Ip":         c.RealIP(),
				"Elapsed":    time.Since(start).Milliseconds(),
				"Method":     c.Request().Method,
				"StatusCode": status,
				"Mode":       c.Request().Header.Values("X-Mode"),
			}
			if rv := c.Get("req"); rv != nil {
				if req, ok := rv.(gateway.HttpRequester); ok {
					if userId := req.GetCurrentAccountId(); userId > 0 {
						meta["UserId"] = userId
					}
				}
			}
			switch {
			case status < 400:
				line.With(meta).DebugF("")
			case status < 500:
				if status == http.StatusFailedDependency {
					line.With(meta).CriticalF("StatusFailedDependency")
				} else {
					line.With(meta).InfoF("")
				}
			default:
				line.With(meta).CriticalF("")
			}
			return nil
		}
	}
}
