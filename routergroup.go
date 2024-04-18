package echoserver

import (
	"github.com/aliworkshop/gateway/v2"
	"github.com/labstack/echo/v4"
	ew "github.com/labstack/echo/v4/middleware"
	"net/http"
	"path/filepath"
)

type routerGroup struct {
	router
	engine      *echo.Echo
	routerGroup *echo.Group
	c           gateway.Controller

	mConfig middlewareConfig
}

func newRouterGroup(e *echo.Echo, c gateway.Controller, config config, path string) *routerGroup {
	r := &routerGroup{
		router: router{
			config: config,
		},
		engine:      e,
		c:           c,
		routerGroup: e.Group(path),
	}
	return r
}

func (r *routerGroup) READ(path string, handlers ...gateway.Handler) {
	hf, mfs := r.match(r.c, handlers...)
	r.routerGroup.GET(path, hf, mfs...)
}
func (r *routerGroup) CREATE(path string, handlers ...gateway.Handler) {
	hf, mfs := r.match(r.c, handlers...)
	r.routerGroup.POST(path, hf, mfs...)
}
func (r *routerGroup) UPDATE(path string, handlers ...gateway.Handler) {
	hf, mfs := r.match(r.c, handlers...)
	r.routerGroup.PUT(path, hf, mfs...)
}
func (r *routerGroup) DELETE(path string, handlers ...gateway.Handler) {
	hf, mfs := r.match(r.c, handlers...)
	r.routerGroup.DELETE(path, hf, mfs...)
}

func (r *routerGroup) STATIC(path string) {
	r.routerGroup.Use(ew.Static(filepath.Join(path)))
}

func (r *routerGroup) ServeHttp(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

func (r *routerGroup) Group(relativePath string) gateway.RouterGroupModel {
	group := &routerGroup{
		router: router{
			config: r.config,
		},
		engine:      r.engine,
		routerGroup: r.routerGroup.Group(relativePath),
		c:           r.c,
	}
	return group
}

func (r *routerGroup) Middleware(handlers ...gateway.Handler) {
	mfs := r.matchMiddleware(r.c, handlers...)
	r.routerGroup.Use(mfs...)
}
