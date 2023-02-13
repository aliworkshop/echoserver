package echoserver

import (
	"context"
	"fmt"
	"github.com/aliworkshop/configlib"
	"github.com/aliworkshop/handlerlib"
	"github.com/aliworkshop/loggerlib"
	"github.com/labstack/echo/v4"
	ew "github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"

	"github.com/aliworkshop/handlerlib/middleware"
	"github.com/aliworkshop/loggerlib/logger"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type echoServer struct {
	router
	server         *echo.Echo
	httpServer     http.Server
	config         config
	configRegistry configlib.Registry

	monitoring handlerlib.MonitoringModel
}

func NewServer(configRegistry configlib.Registry) handlerlib.ServerModel {
	var cfg config
	if err := configRegistry.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	cfg.Initialize()
	gs := &echoServer{
		router: router{
			config: cfg,
		},
		config:         cfg,
		configRegistry: configRegistry,
		monitoring:     handlerlib.DefaultMonitoring,
	}
	s := echo.New()
	s.Use(ew.Recover())
	if gs.config.Http.Development {
		s.Use(ew.Logger())
	} else {
		l, err := loggerlib.GetLogger(configRegistry.ValueOf("http.logger"))
		if err != nil {
			panic("logger for http is not set. set http server config to development")
		}
		s.Use(NewLoggerHandler(l, gs.config.Http))
	}
	s.Use(ew.CORSWithConfig(ew.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderAuthorization, echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept,
		},
	}))
	gs.server = s

	return gs
}

func (gs *echoServer) SetMonitoringHandler(monitoring handlerlib.MonitoringModel) {
	gs.monitoring = monitoring
}

func (gs *echoServer) SetupMiddlewares(logger logger.Logger, languageBundle *i18n.Bundle) {
	middlewares := make([]handlerlib.HandlerModel, 0)
	for key, h := range gs.config.Middlewares {
		handler := NewHandlerModel(logger, languageBundle)
		m := middleware.Get(handler,
			gs.configRegistry.
				ValueOf("middlewares").
				ValueOf(key),
			h.Type)
		if m == nil {
			panic(fmt.Sprintf("could not find middleware for type: %v", h.Type))
		}
		middlewares = append(middlewares, m)
	}
	gs.Middleware(middlewares...)
}

func (gs *echoServer) Middleware(handlers ...handlerlib.HandlerModel) {
	_, mfs := gs.match(gs.monitoring, handlers...)
	gs.server.Use(mfs...)
}

func (gs *echoServer) Run(addr ...string) error {
	if addr == nil || len(addr) == 0 {
		addr = []string{"127.0.0.1:8080"}
	}
	err := gs.server.Start(addr[0])
	if err != nil {
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
	return nil
}

func (gs *echoServer) NewRouterGroup(path string) handlerlib.RouterGroupModel {
	rg := newRouterGroup(gs.server, gs.config, path)
	rg.monitoring = gs.monitoring
	return rg
}

func (gs *echoServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return gs.server.Shutdown(ctx)
}
