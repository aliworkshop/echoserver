package echoserver

import (
	"context"
	"github.com/aliworkshop/dfilter"
	errors "github.com/aliworkshop/error"
	"github.com/aliworkshop/gateway/v2"
	"github.com/aliworkshop/gateway/v2/authorization"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"
)

type request struct {
	uid               string
	context           echo.Context
	connectionContext context.Context
	auth              authorization.Authorizer
	body              any
	filters           map[string][]string
	language          gateway.Language
	responded         bool
	paginator         gateway.Paginator
	dFilters          []dfilter.Filter

	temp    map[string]interface{}
	tempMtx *sync.Mutex
}

func NewRequest(ctx echo.Context, languageBundle *i18n.Bundle) gateway.Requester {
	req := &request{
		temp:    make(map[string]interface{}),
		tempMtx: new(sync.Mutex),
	}
	req.SetContext(ctx)
	if languageBundle != nil {
		acceptLanguage := ctx.Request().Header.Get("Accept-Language")
		req.SetLanguage(gateway.NewLanguage(languageBundle, acceptLanguage))
	}

	uid := ctx.Request().Header.Get("X-Request-UID")
	if uid == "" {
		uid = uuid.New().String()
	}
	req.SetUid(uid)
	req.connectionContext = context.WithValue(ctx.Request().Context(), "uid", uid)
	return req
}

func (r *request) SetUid(uid string) {
	r.uid = uid
	r.context.Set("UID", uid)
}

func (r *request) GetUid() string {
	return r.uid
}

func (r *request) SetConnectionContext(ctx context.Context) {
	r.connectionContext = ctx
}

func (r *request) GetConnectionContext() context.Context {
	return r.connectionContext
}

func (r *request) GetContext() interface{} {
	return r.context
}

func (r *request) GetClientIp() string {
	return r.context.RealIP()
}

func (r *request) GetMethod() string {
	return r.context.Request().Method
}

func (r *request) GetPath() string {
	return r.context.Path()
}

func (r *request) GetHeader(key string) string {
	return r.context.Request().Header.Get(key)
}

func (r *request) Cookie(name string) (string, error) {
	cookie, err := r.context.Request().Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (r *request) SetCookie(cookie interface{}) {
	http.SetCookie(r.context.Response(), cookie.(*http.Cookie))
}

func (r *request) SetBody(i interface{}) {
	r.body = i
}

func (r *request) GetBody() interface{} {
	return r.body
}

func (r *request) BindRequest(body interface{}) errors.ErrorModel {
	err := r.context.Bind(body)
	if err != nil {
		return errors.Validation(err)
	}
	if err = r.context.Validate(body); err != nil {
		return errors.Validation(err).WithMessage(err.Error())
	}
	r.body = body
	return nil
}

func (r *request) SetLanguage(language gateway.Language) {
	r.language = language
}

func (r *request) MustLocalize(lc *i18n.LocalizeConfig) string {
	result, err := r.language.Localize(lc)
	if err != nil {
		log.Fatalf("error on localize, err: %v", err)
	}
	return result
}

func (r *request) ShouldLocalize(lc *i18n.LocalizeConfig) string {
	result, err := r.language.Localize(lc)
	if err != nil {
		log.Printf("error on localize, err: %v", err)
	}
	return result
}

func (r *request) Localize(msgId string, message string, params ...map[string]interface{}) string {
	if r.language == nil {
		return message
	}
	var p map[string]interface{}
	if params != nil && len(params) > 0 {
		p = params[0]
	}
	msg, err := r.language.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    msgId,
			Other: message,
		},
		TemplateData: p,
	})
	if err != nil {
		log.Printf("error on localize, err: %v", err)
	}
	return msg
}

func (r *request) SetContext(i interface{}) {
	r.context = i.(echo.Context)
}

func (r *request) GetParam(key string) string {
	return r.context.Param(key)
}

func (r *request) GetQuery(key string) string {
	return r.context.QueryParam(key)
}

func (r *request) GetFile(key string) (*multipart.FileHeader, error) {
	return r.context.FormFile(key)
}

func (r *request) GetFiles(key string) ([]*multipart.FileHeader, error) {
	form, err := r.context.MultipartForm()
	if err != nil {
		return nil, err
	}
	return form.File[key], nil
}

func (r *request) Paginator() gateway.Paginator {
	if r.paginator != nil {
		return r.paginator
	}
	p := NewPaginator()
	page, _ := strconv.Atoi(r.GetQuery("$page"))
	p.SetPage(page)

	limit, _ := strconv.Atoi(r.GetQuery("$limit"))
	p.SetLimit(limit)

	p.SetSort(r.GetQuery("$sortby"))
	r.paginator = p
	return p
}

func (r *request) Request() *http.Request {
	return r.context.Request()
}

func (r *request) Writer() http.ResponseWriter {
	return r.context.Response()
}

func (r *request) IsResponded() bool {
	return r.responded
}

func (r *request) SetResponded(responded bool) {
	r.responded = responded
}

func (r *request) SetDynamicFilters(fs []dfilter.Filter) {
	r.dFilters = fs
}

func (r *request) GetDynamicFilters() []dfilter.Filter {
	return r.dFilters
}

func (r *request) Filters() map[string][]string {
	if r.filters != nil {
		return r.filters
	}
	r.filters = make(map[string][]string)
	for k, v := range r.context.Request().URL.Query() {
		r.filters[k] = v
	}
	return r.filters
}

func (r *request) Token() (token string) {
	if r.auth == nil {
		return
	}
	return r.auth.Token()
}

func (r *request) SetAuth(auth authorization.Authorizer) {
	r.auth = auth
}

func (r *request) GetAuth() authorization.Authorizer {
	return r.auth
}

func (r *request) GetCurrentAccountId() uint64 {
	auth := r.GetAuth()
	if auth != nil {
		return auth.GetCurrentAccountId()
	}
	return 0
}

func (r *request) GetCurrentAccountUuid() string {
	auth := r.GetAuth()
	if auth != nil {
		return auth.GetCurrentAccountUuid()
	}
	return ""
}

func (r *request) IsAuthenticated() bool {
	if r.auth == nil {
		return false
	}
	return r.auth.IsAuthenticated()
}

func (r *request) GetScopes() []string {
	if r.auth == nil || !r.auth.IsAuthenticated() {
		return nil
	}
	return r.auth.GetScopes()
}

func (r *request) HasScope(scopes ...string) bool {
	if r.auth == nil || !r.auth.IsAuthenticated() {
		return false
	}
	return r.auth.HasScope(scopes...)
}

func (r *request) SetTemp(key string, value interface{}) {
	r.tempMtx.Lock()
	r.temp[key] = value
	r.tempMtx.Unlock()
}

func (r *request) GetTemp(key string) interface{} {
	r.tempMtx.Lock()
	temp, _ := r.temp[key]
	r.tempMtx.Unlock()
	return temp
}

func (r *request) GetStatusCode() int {
	return r.context.Response().Status
}

func (r *request) Websocket() (gateway.WebSocketHandler, errors.ErrorModel) {
	ws, err := upgrade(r.context)
	if err != nil {
		return nil, errors.HandleError(err)
	}
	return ws, nil
}

func (r *request) RespondBlob(status gateway.Status, contentType string, body []byte) errors.ErrorModel {
	err := r.context.Blob(getStatusCode(status), contentType, body)
	if err != nil {
		return errors.HandleError(err)
	}
	return nil
}

func (r *request) RespondStream(status gateway.Status, contentType string, reader io.Reader) errors.ErrorModel {
	err := r.context.Stream(getStatusCode(status), contentType, reader)
	if err != nil {
		return errors.HandleError(err)
	}
	return nil
}

func (r *request) RespondFile(file string) errors.ErrorModel {
	err := r.context.File(file)
	if err != nil {
		return errors.HandleError(err)
	}
	return nil
}

func (r *request) RespondFsFile(file string, filesystem fs.FS) errors.ErrorModel {
	err := r.context.File(file)
	if err != nil {
		return errors.HandleError(err)
	}
	return nil
}
