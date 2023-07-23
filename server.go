package echoserver

import (
	"context"
	"fmt"
	"github.com/aliworkshop/configer"
	errors "github.com/aliworkshop/error"
	"github.com/aliworkshop/gateway/v2"
	"github.com/aliworkshop/gateway/v2/middleware"
	"github.com/aliworkshop/logger"
	"github.com/go-playground/validator/v10"
	echop "github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	ew "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"path/filepath"
	"time"
)

type echoServer struct {
	router
	server         *echo.Echo
	httpServer     http.Server
	config         config
	configRegistry configer.Registry
	controller     gateway.Controller

	monitoring gateway.MonitoringModel
}

func NewServer(configRegistry configer.Registry) gateway.ServerModel {
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
		monitoring:     gateway.DefaultMonitoring,
	}
	s := echo.New()
	if !cfg.Development {
		s.Use(ew.Recover())
	}

	group := s.Group("assets")
	group.Use(ew.Static(filepath.Join("uploads")))

	if gs.config.Http.Development {
		s.Use(ew.Logger())
	} else {
		l, err := logger.GetLogger(configRegistry.ValueOf("http.logger"))
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
	s.Validator = &customValidator{validator: validator.New()}
	gs.server = s

	return gs
}

func (gs *echoServer) SetMonitoringHandler(monitoring gateway.MonitoringModel) {
	gs.monitoring = monitoring
}

func (gs *echoServer) AddMonitoring(m *gateway.Monitoring) (prometheus.Collector, errors.ErrorModel) {
	metric := echop.NewMetric(&echop.Metric{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Type:        m.Type.String(),
		Args:        m.Args,
		Buckets:     m.Buckets,
	}, m.Subsystem)
	if err := prometheus.Register(metric); err != nil {
		return nil, errors.New(fmt.Errorf("%s could not be registered in Prometheus: %v", m.Name, err))
	}
	return metric, nil
}

func (gs *echoServer) StartMonitoring() {
	p := echop.NewPrometheus("app", nil)
	p.MetricsPath = "monitoring/metrics"
	p.Use(gs.server)
}

func (gs *echoServer) SetupMiddlewares() {
	middlewares := make([]gateway.Handler, 0)
	for key, h := range gs.config.Middlewares {
		m := middleware.Get(gs.configRegistry.
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

func (gs *echoServer) Middleware(handlers ...gateway.Handler) {
	_, mfs := gs.match(gs.monitoring, gs.controller, handlers...)
	gs.server.Use(mfs...)
}

func (gs *echoServer) SetController(controller gateway.Controller) {
	gs.controller = controller
}

func (gs *echoServer) GetController() gateway.Controller {
	return gs.controller
}

func (gs *echoServer) NewRouterGroup(path string) gateway.RouterGroupModel {
	rg := newRouterGroup(gs.server, gs.controller, gs.config, path)
	rg.monitoring = gs.monitoring
	return rg
}

func (gs *echoServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return gs.server.Shutdown(ctx)
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
