package echoserver

import (
	"fmt"
	"github.com/aliworkshop/configer"
	"github.com/aliworkshop/gateway/v2"
	"github.com/aliworkshop/gateway/v2/middleware"
	"github.com/labstack/echo/v4"
	ew "github.com/labstack/echo/v4/middleware"
	"path/filepath"
)

type routerGroup struct {
	router
	engine      *echo.Echo
	routerGroup *echo.Group
	c           gateway.Controller

	mConfig middlewareConfig

	monitoring gateway.MonitoringModel
}

func newRouterGroup(e *echo.Echo, c gateway.Controller, config config, path string) *routerGroup {
	r := &routerGroup{
		router: router{
			config: config,
		},
		engine:      e,
		c:           c,
		routerGroup: e.Group(path),
		monitoring:  gateway.DefaultMonitoring,
	}
	return r
}

func (r *routerGroup) SetMonitoringHandler(monitoring gateway.MonitoringModel) {
	r.monitoring = monitoring
}

func (r *routerGroup) READ(path string, handlers ...gateway.Handler) {
	hf, mfs := r.match(r.monitoring, r.c, handlers...)
	r.routerGroup.GET(path, hf, mfs...)
}
func (r *routerGroup) CREATE(path string, handlers ...gateway.Handler) {
	hf, mfs := r.match(r.monitoring, r.c, handlers...)
	r.routerGroup.POST(path, hf, mfs...)
}
func (r *routerGroup) UPDATE(path string, handlers ...gateway.Handler) {
	hf, mfs := r.match(r.monitoring, r.c, handlers...)
	r.routerGroup.PUT(path, hf, mfs...)
}
func (r *routerGroup) DELETE(path string, handlers ...gateway.Handler) {
	hf, mfs := r.match(r.monitoring, r.c, handlers...)
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

func (r *routerGroup) SetupMiddlewares(registry configer.Registry) {
	if err := registry.Unmarshal(&r.mConfig); err != nil {
		panic(err)
	}
	for key, h := range r.mConfig.Middlewares {
		m := middleware.Get(registry.
			ValueOf("middlewares").
			ValueOf(key),
			h.Type)
		if m == nil {
			panic(fmt.Sprintf("could not find middleware for type: %v", h.Type))
		}
		r.Middleware(m)
	}
}

func (r *routerGroup) Middleware(handlers ...gateway.Handler) {
	mfs := r.matchMiddleware(r.monitoring, r.c, handlers...)
	r.routerGroup.Use(mfs...)
}
