package echoserver

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type CSRFConfig struct {
	CookieKey string
	HeaderKey string
}

type middlewareConfig struct {
	Middlewares map[string]struct {
		Type   string
		Config interface{}
	}
}

type Http struct {
	Development bool
	Logger      struct {
		SkipPaths []string
	}
	ConnectionTimeout time.Duration
	ServiceName       string `mapstructure:"servicename"`
	CSRF              struct {
		SessionTypes map[string]*CSRFConfig
	}
	Cors struct {
		AllowOrigins []string
		AllowMethods []string
		AllowHeaders []string
	}
}

type config struct {
	middlewareConfig
	Http
}

func (c *config) Initialize() {
	if c.ConnectionTimeout == 0 {
		c.ConnectionTimeout = time.Second * 30
	}
	if c.CSRF.SessionTypes == nil {
		c.CSRF.SessionTypes = map[string]*CSRFConfig{
			"DEFAULT": {
				CookieKey: "CSRF_TOKEN",
				HeaderKey: "X-CSRF-TOKEN",
			},
			"NORMAL": {
				CookieKey: "CSRF_TOKEN_NORMAL",
				HeaderKey: "X-CSRF-TOKEN-NORMAL",
			},
			"IMPORTANT": {
				CookieKey: "CSRF_TOKEN_IMPORTANT",
				HeaderKey: "X-CSRF-TOKEN-IMPORTANT",
			},
		}
	}
	if len(c.Cors.AllowOrigins) == 0 {
		c.Cors.AllowOrigins = []string{"*"}
	}
	if len(c.Cors.AllowMethods) == 0 {
		c.Cors.AllowMethods = []string{
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions,
		}
	}
	if len(c.Cors.AllowHeaders) == 0 {
		c.Cors.AllowHeaders = []string{
			echo.HeaderAuthorization, echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept,
		}
	}
}
