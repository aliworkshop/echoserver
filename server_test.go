package echoserver

import (
	"encoding/json"
	"github.com/aliworkshop/logger/writers"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aliworkshop/errors"
	"github.com/aliworkshop/gateway/v2"
	"github.com/aliworkshop/logger"
)

type stubLogger struct{}

func (s *stubLogger) Clone() logger.Logger             { return s }
func (s *stubLogger) With(logger.Field) logger.Logger  { return s }
func (s *stubLogger) WithId(string) logger.Logger      { return s }
func (s *stubLogger) WithUid(string) logger.Logger     { return s }
func (s *stubLogger) WithSource(string) logger.Logger  { return s }
func (s *stubLogger) DebugF(string, ...interface{})    {}
func (s *stubLogger) InfoF(string, ...interface{})     {}
func (s *stubLogger) WarnF(string, ...interface{})     {}
func (s *stubLogger) ErrorF(string, ...interface{})    {}
func (s *stubLogger) CriticalF(string, ...interface{}) {}
func (s *stubLogger) FatalF(string, ...interface{})    {}

type handlerFunc func(req gateway.HttpRequester) (any, errors.ErrorModel)

func (h handlerFunc) Handle(req gateway.HttpRequester) (any, errors.ErrorModel) { return h(req) }

func newTestRouter(t *testing.T, path string) (gateway.RouterGroupModel, gateway.ServerModel) {
	t.Helper()

	controller := gateway.NewController(NewResponder(nil), logger.NewSimpleLogger(writers.DebugLevel, logger.JsonEncoding))
	server := NewTestServer(controller)
	return server.NewRouterGroup(path), server
}

func TestServer_READ_OK(t *testing.T) {
	rg, _ := newTestRouter(t, "/api")
	rg.READ("/ping", handlerFunc(func(req gateway.HttpRequester) (any, errors.ErrorModel) {
		return map[string]bool{"pong": true}, nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()
	rg.ServeHttp(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200; body=%s", rec.Code, rec.Body.String())
	}

	var resp gateway.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, rec.Body.String())
	}
	items, ok := resp.Items.(map[string]any)
	if !ok {
		t.Fatalf("Items wrong type: %T", resp.Items)
	}
	if pong, _ := items["pong"].(bool); !pong {
		t.Fatalf("expected pong=true; got %v", items)
	}
	if got := rec.Header().Get("X-Request-Uuid"); got == "" {
		t.Errorf("expected X-Request-Uuid header to be set")
	}
}

func TestServer_CREATE_201(t *testing.T) {
	rg, _ := newTestRouter(t, "/api")
	rg.CREATE("/widgets", handlerFunc(func(req gateway.HttpRequester) (any, errors.ErrorModel) {
		return map[string]any{"id": uint64(42)}, nil
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/widgets", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rg.ServeHttp(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d; want 201; body=%s", rec.Code, rec.Body.String())
	}
}

func TestServer_NotFoundError_404(t *testing.T) {
	rg, _ := newTestRouter(t, "/api")
	rg.READ("/missing", handlerFunc(func(req gateway.HttpRequester) (any, errors.ErrorModel) {
		return nil, errors.NotFound()
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/missing", nil)
	rec := httptest.NewRecorder()
	rg.ServeHttp(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want 404; body=%s", rec.Code, rec.Body.String())
	}
}

func TestServer_ValidationError_400(t *testing.T) {
	rg, _ := newTestRouter(t, "/api")
	rg.CREATE("/widgets", handlerFunc(func(req gateway.HttpRequester) (any, errors.ErrorModel) {
		return nil, errors.Validation()
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/widgets", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	rg.ServeHttp(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestServer_DELETE_NoBody_204(t *testing.T) {
	rg, _ := newTestRouter(t, "/api")
	rg.DELETE("/widgets/:id", handlerFunc(func(req gateway.HttpRequester) (any, errors.ErrorModel) {
		return nil, nil
	}))

	req := httptest.NewRequest(http.MethodDelete, "/api/widgets/1", nil)
	rec := httptest.NewRecorder()
	rg.ServeHttp(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want 204; body=%s", rec.Code, rec.Body.String())
	}
}

func TestServer_PaginationQueryParams(t *testing.T) {
	rg, _ := newTestRouter(t, "/api")
	var seenPage, seenSize int
	rg.READ("/items", handlerFunc(func(req gateway.HttpRequester) (any, errors.ErrorModel) {
		p := req.Paginator()
		seenPage = p.GetPage()
		seenSize = p.GetPageSize()
		p.SetTotal(123)
		return []int{1, 2, 3}, nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/items?$page=2&$page_size=25", nil)
	rec := httptest.NewRecorder()
	rg.ServeHttp(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200; body=%s", rec.Code, rec.Body.String())
	}
	if seenPage != 2 || seenSize != 25 {
		t.Fatalf("paginator parse: page=%d size=%d; want 2/25", seenPage, seenSize)
	}
}
