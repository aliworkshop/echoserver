package echoserver

import (
	"fmt"
	"github.com/aliworkshop/configlib"
	"github.com/aliworkshop/gateway"
	"github.com/aliworkshop/gateway/middleware"
	"github.com/aliworkshop/loggerlib/logger"
	"github.com/labstack/echo/v4"
	ew "github.com/labstack/echo/v4/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"path/filepath"
)

type routerGroup struct {
	router
	engine      *echo.Echo
	routerGroup *echo.Group

	mConfig middlewareConfig

	monitoring gateway.MonitoringModel
}

func newRouterGroup(e *echo.Echo, config config, path string) *routerGroup {
	r := &routerGroup{
		router: router{
			config: config,
		},
		engine:      e,
		routerGroup: e.Group(path),
		monitoring:  gateway.DefaultMonitoring,
	}
	return r
}

func (r *routerGroup) SetMonitoringHandler(monitoring gateway.MonitoringModel) {
	r.monitoring = monitoring
}

func (r *routerGroup) READ(path string, handlers ...gateway.HandlerEngine) {
	hf, mfs := r.match(r.monitoring, handlers...)
	r.routerGroup.GET(path, hf, mfs...)
}
func (r *routerGroup) CREATE(path string, handlers ...gateway.HandlerEngine) {
	hf, mfs := r.match(r.monitoring, handlers...)
	r.routerGroup.POST(path, hf, mfs...)
}
func (r *routerGroup) UPDATE(path string, handlers ...gateway.HandlerEngine) {
	hf, mfs := r.match(r.monitoring, handlers...)
	r.routerGroup.PUT(path, hf, mfs...)
}
func (r *routerGroup) DELETE(path string, handlers ...gateway.HandlerEngine) {
	hf, mfs := r.match(r.monitoring, handlers...)
	r.routerGroup.DELETE(path, hf, mfs...)
}

func (r *routerGroup) STATIC(path string) {
	r.routerGroup.Use(ew.Static(filepath.Join(path)))
}

func (r *routerGroup) Group(relativePath string) gateway.RouterGroupModel {
	group := &routerGroup{
		router: router{
			config: r.config,
		},
		engine:      r.engine,
		routerGroup: r.routerGroup.Group(relativePath),
		monitoring:  r.monitoring,
	}
	return group
}

func (r *routerGroup) SetupMiddlewares(registry configlib.Registry,
	logger logger.Logger, languageBundle *i18n.Bundle) {
	if err := registry.Unmarshal(&r.mConfig); err != nil {
		panic(err)
	}
	for key, h := range r.mConfig.Middlewares {
		handler := NewHandlerEngine(logger, languageBundle)
		m := middleware.Get(handler,
			registry.
				ValueOf("middlewares").
				ValueOf(key),
			h.Type)
		if m == nil {
			panic(fmt.Sprintf("could not find middleware for type: %v", h.Type))
		}
		r.Middleware(m)
	}
}

func (r *routerGroup) Middleware(handlers ...gateway.HandlerEngine) {
	_, mfs := r.match(r.monitoring, handlers...)
	r.routerGroup.Use(mfs...)
}
