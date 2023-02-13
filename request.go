package echoserver

import (
	"context"
	"github.com/aliworkshop/errorslib"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"sync"

	"github.com/aliworkshop/dfilterlib"
	"github.com/aliworkshop/handlerlib"
	"github.com/aliworkshop/handlerlib/authorization"
	"github.com/aliworkshop/loggerlib/logger"
	"github.com/google/uuid"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type request struct {
	uid               string
	context           echo.Context
	connectionContext context.Context
	isUpgraded        bool
	auth              authorization.Model
	body              interface{}
	filters           map[string][]string
	language          handlerlib.Language
	log               logger.Logger
	responded         bool
	*paginator
	dFilters []dfilterlib.Filter

	temp    map[string]interface{}
	tempMtx *sync.Mutex
}

func NewRequest(ctx echo.Context, languageBundle *i18n.Bundle) handlerlib.RequestModel {
	req := &request{
		temp:    make(map[string]interface{}),
		tempMtx: new(sync.Mutex),
	}
	req.SetContext(ctx)
	if languageBundle != nil {
		acceptLanguage := ctx.Request().Header.Get("Accept-Language")
		req.language = handlerlib.Language{
			AcceptLanguage: acceptLanguage,
			Localizer:      i18n.NewLocalizer(languageBundle, acceptLanguage, "EN"),
		}
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

func (r *request) Logger() logger.Logger {
	return r.log
}

func (r *request) WithLogger(l logger.Logger) handlerlib.RequestModel {
	r.log = l.With(logger.Field{"UID": r.uid})
	return r
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

func (r *request) HandleRequestBody(body interface{}) errorslib.ErrorModel {
	err := r.context.Bind(body)
	if err != nil {
		return errorslib.Validation(err)
	}
	r.body = body
	return nil
}

func (r *request) HandleRequestJsonBody(body interface{}) errorslib.ErrorModel {
	err := r.context.Bind(body)
	if err != nil {
		return errorslib.Validation(err)
	}
	r.body = body
	return nil
}

func (r *request) HandleRequestParams(params interface{}) (err error) {
	err = r.context.Bind(params)
	if err != nil {
		return
	}
	return nil
}

func (r *request) GetLanguage() handlerlib.Language {
	return r.language
}

func (r *request) MustLocalize(lc *i18n.LocalizeConfig) string {
	return r.language.Localizer.MustLocalize(lc)
}

func (r *request) ShouldLocalize(lc *i18n.LocalizeConfig) string {
	result, err := r.language.Localizer.Localize(lc)
	if err != nil {
		r.Logger().DebugF("error on localize, err: %v", err)
	}
	return result
}

func (r *request) Localize(msgId string, message string, params ...map[string]interface{}) string {
	if r.language.Localizer == nil {
		return message
	}
	var p map[string]interface{}
	if params != nil && len(params) > 0 {
		p = params[0]
	}
	msg, err := r.language.Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    msgId,
			Other: message,
		},
		TemplateData: p,
	})
	if err != nil {
		r.Logger().DebugF("error on localize, err: %v", err)
	}
	return msg
}

func (r *request) SetContext(i interface{}) {
	r.context = i.(echo.Context)
}

func (r *request) SetModel(i interface{}) {
	r.body = i
}

func (r *request) GetParam(key string) (interface{}, bool) {
	if param := r.context.Param(key); param != "" {
		return param, true
	}
	return "", false
}

func (r *request) GetQuery(key string) (string, bool) {
	if query := r.context.QueryParam(key); query != "" {
		return query, true
	}
	return "", false
}

func (r *request) Paging() handlerlib.Pagination {
	if r.paginator != nil {
		return r.paginator
	}
	p := new(paginator)
	page := r.context.QueryParam("$page")
	p.page, _ = strconv.Atoi(page)

	pp := r.context.QueryParam("$perpage")
	p.perPage, _ = strconv.Atoi(pp)

	p.sortBy = r.context.QueryParam("$sortby")
	r.paginator = p
	return p
}

func (r request) BaseRequest() *http.Request {
	return r.context.Request()
}

func (r request) BaseWriter() http.ResponseWriter {
	return r.context.Response()
}

func (r *request) IsResponded() bool {
	return r.responded
}

func (r *request) SetResponded(responded bool) {
	r.responded = responded
}

func (r *request) SetDynamicFilters(fs []dfilterlib.Filter) {
	r.dFilters = fs
}

func (r *request) GetDynamicFilters() []dfilterlib.Filter {
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

// authorization methods
func (r *request) Token() (token string) {
	if r.auth == nil {
		return
	}
	return r.auth.Token()
}

func (r *request) SetAuth(auth authorization.Model) {
	r.auth = auth
}

func (r *request) GetAuth() authorization.Model {
	return r.auth
}

func (r *request) GetCurrentAccountId() interface{} {
	auth := r.GetAuth()
	if auth != nil {
		return auth.GetCurrentAccountId()
	}
	return nil
}

func (r *request) IsAuthenticated() bool {
	if r.auth == nil {
		return false
	}
	return r.auth.IsAuthenticated()
}

func (r *request) GetScopes() []string {
	if r.auth == nil {
		return nil
	}
	return r.auth.GetScopes()
}

func (r *request) HasScope(scopes ...string) bool {
	if r.auth == nil {
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
