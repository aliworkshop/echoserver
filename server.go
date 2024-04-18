package echoserver

import (
	"context"
	"fmt"
	"github.com/aliworkshop/configer"
	errors "github.com/aliworkshop/error"
	"github.com/aliworkshop/gateway/v2"
	"github.com/aliworkshop/logger"
	"github.com/go-playground/validator/v10"
	echop "github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	ew "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type echoServer struct {
	router
	server         *echo.Echo
	config         config
	configRegistry configer.Registry
	controller     gateway.Controller
}

func NewServer(configRegistry configer.Registry) gateway.ServerModel {
	var cfg config
	if err := configRegistry.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	cfg.Initialize()
	es := &echoServer{
		router: router{
			config: cfg,
		},
		config:         cfg,
		configRegistry: configRegistry,
	}
	s := echo.New()
	if !cfg.Development {
		s.Use(ew.Recover())
	}

	if es.config.Http.Development {
		s.Use(ew.Logger())
	} else {
		l, err := logger.GetLogger(configRegistry.ValueOf("http.logger"))
		if err != nil {
			panic("logger for http is not set. set http server config to development")
		}
		s.Use(NewLoggerHandler(l, es.config.Http))
	}
	s.Use(ew.CORSWithConfig(ew.CORSConfig{
		AllowOrigins: cfg.Cors.AllowOrigins,
		AllowMethods: cfg.Cors.AllowMethods,
		AllowHeaders: cfg.Cors.AllowHeaders,
	}))
	s.Validator = &customValidator{validator: validator.New()}
	es.server = s

	return es
}

func NewTestServer(c gateway.Controller) gateway.ServerModel {
	return &echoServer{
		server:     echo.New(),
		controller: c,
	}
}

func (es *echoServer) AddMonitoring(m *gateway.Monitoring) (prometheus.Collector, errors.ErrorModel) {
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

func (es *echoServer) StartMonitoring() {
	p := echop.NewPrometheus("app", func(c echo.Context) bool {
		if strings.HasSuffix(c.Path(), "monitoring/metrics") {
			return true
		}
		return false
	})
	p.MetricsPath = "monitoring/metrics"
	p.Use(es.server)
}

func (es *echoServer) Middleware(handlers ...gateway.Handler) {
	_, mfs := es.match(es.controller, handlers...)
	es.server.Use(mfs...)
}

func (es *echoServer) SetController(controller gateway.Controller) {
	es.controller = controller
}

func (es *echoServer) GetController() gateway.Controller {
	return es.controller
}

func (es *echoServer) NewRouterGroup(path string) gateway.RouterGroupModel {
	rg := newRouterGroup(es.server, es.controller, es.config, path)
	return rg
}

func (es *echoServer) LoadHtml(path string) {
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob(path)),
	}
	es.server.Renderer = renderer
}

func (es *echoServer) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return es.server.Shutdown(ctx)
}

func (es *echoServer) Run(addr ...string) error {
	if addr == nil || len(addr) == 0 {
		addr = []string{"127.0.0.1:8080"}
	}
	err := es.server.Start(addr[0])
	if err != nil {
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
	return nil
}
