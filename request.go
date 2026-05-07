package echoserver

import (
	"context"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"

	ad "github.com/aliworkshop/authorizer/port"
	"github.com/aliworkshop/dfilter"
	"github.com/aliworkshop/errors"
	"github.com/aliworkshop/gateway/v2"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type request struct {
	uid               string
	requestUUID       string
	context           echo.Context
	connectionContext context.Context
	auth              ad.Authorizer
	body              any
	filters           map[string][]string
	language          gateway.Language
	responded         bool
	paginator         gateway.IPaginator
	sorter            gateway.Sorter
	dFilters          []dfilter.Filter
	requestScopes     []string

	temp    map[string]any
	tempMtx sync.Mutex
}

func NewRequest(ctx echo.Context, languageBundle *i18n.Bundle) gateway.HttpRequester {
	r := &request{
		temp: make(map[string]any),
	}
	r.SetContext(ctx)
	if languageBundle != nil {
		acceptLanguage := ctx.Request().Header.Get("Accept-Language")
		r.SetLanguage(gateway.NewLanguage(languageBundle, acceptLanguage))
	}

	uid := ctx.Request().Header.Get("X-Request-UID")
	if uid == "" {
		uid = uuid.New().String()
	}
	r.SetUid(uid)
	r.requestUUID = uid
	r.connectionContext = context.WithValue(ctx.Request().Context(), "uid", uid)
	return r
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

func (r *request) GetContext() any {
	return r.context
}

func (r *request) SetContext(i any) {
	r.context = i.(echo.Context)
}

func (r *request) GetHttpContext() any {
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

func (r *request) SetCookie(cookie any) {
	http.SetCookie(r.context.Response(), cookie.(*http.Cookie))
}

func (r *request) SetBody(i any) {
	r.body = i
}

func (r *request) GetBody() any {
	return r.body
}

func (r *request) BindRequest(body gateway.Validatable) errors.ErrorModel {
	if err := r.context.Bind(body); err != nil {
		return errors.Validation(err).WithProperty("error", err.Error())
	}
	return body.Validate(getValidator(r.context), r.language)
}

func (r *request) SetLanguage(language gateway.Language) {
	r.language = language
}

func (r *request) GetLanguage() gateway.Language {
	return r.language
}

func (r *request) MustLocalize(lc *i18n.LocalizeConfig) string {
	if r.language == nil {
		if lc.DefaultMessage != nil {
			return lc.DefaultMessage.Other
		}
		return ""
	}
	result, err := r.language.Localize(lc)
	if err != nil {
		log.Fatalf("error on localize, err: %v", err)
	}
	return result
}

func (r *request) ShouldLocalize(lc *i18n.LocalizeConfig) string {
	if r.language == nil {
		if lc.DefaultMessage != nil {
			return lc.DefaultMessage.Other
		}
		return ""
	}
	result, err := r.language.Localize(lc)
	if err != nil {
		log.Printf("error on localize, err: %v", err)
	}
	return result
}

func (r *request) Localize(msgId string, message string, params ...map[string]any) string {
	if r.language == nil {
		return message
	}
	var p map[string]any
	if len(params) > 0 {
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

func (r *request) Translate(msgId, message string, params ...any) string {
	if r.language == nil {
		return message
	}
	var p any
	if len(params) > 0 {
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

func (r *request) GetParam(key string) string {
	return r.context.Param(key)
}

func (r *request) GetQuery(key string) string {
	return r.context.QueryParam(key)
}

func (r *request) FormValue(key string) string {
	return r.context.FormValue(key)
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

func (r *request) GetAllFiles() (map[string][]*multipart.FileHeader, error) {
	form, err := r.context.MultipartForm()
	if err != nil {
		return nil, err
	}
	return form.File, nil
}

func (r *request) Paginator() gateway.IPaginator {
	if r.paginator != nil {
		return r.paginator
	}
	p := gateway.NewPaginator()
	if page, err := strconv.Atoi(r.GetQuery("$page")); err == nil {
		p.SetPage(int32(page))
	}
	if size, err := strconv.Atoi(r.GetQuery("$page_size")); err == nil {
		p.SetPageSize(int32(size))
	}
	r.paginator = p
	return p
}

func (r *request) SetPaginator(p gateway.IPaginator) {
	r.paginator = p
}

func (r *request) Sorter() gateway.Sorter {
	if r.sorter != nil {
		return r.sorter
	}
	s := gateway.NewSorter()
	s.SetSort(r.GetQuery("$sortby"))
	r.sorter = s
	return s
}

func (r *request) SetSorter(s gateway.Sorter) {
	r.sorter = s
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

func (r *request) SetIsResponded(responded bool) {
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

func (r *request) SetAuth(auth ad.Authorizer) {
	r.auth = auth
}

func (r *request) GetAuth() ad.Authorizer {
	return r.auth
}

func (r *request) IsAuthenticated() bool {
	if r.auth == nil {
		return false
	}
	return r.auth.IsAuthenticated()
}

func (r *request) GetCurrentAccountId() uint64 {
	if r.auth == nil {
		return 0
	}
	return r.auth.GetCurrentAccountId()
}

func (r *request) GetCurrentAccountUuid() string {
	if r.auth == nil {
		return ""
	}
	return r.auth.GetCurrentAccountUuid()
}

func (r *request) GetCurrentAccountEmail() string {
	if r.auth == nil || r.auth.GetClaim() == nil {
		return ""
	}
	return r.auth.GetClaim().GetEmail()
}

func (r *request) GetIssuer() string {
	if r.auth == nil || r.auth.GetClaim() == nil {
		return ""
	}
	return r.auth.GetClaim().GetIssuer()
}

func (r *request) GetScopes() map[string]uint16 {
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

func (r *request) GetRequestScopes() []string {
	return r.requestScopes
}

func (r *request) SetRequestScopes(scopes ...string) {
	r.requestScopes = scopes
}

func (r *request) GetRoles() map[string]uint16 {
	if r.auth == nil || !r.auth.IsAuthenticated() {
		return nil
	}
	return r.auth.GetRoles()
}

func (r *request) HasRole(roles ...string) bool {
	if r.auth == nil || !r.auth.IsAuthenticated() {
		return false
	}
	return r.auth.HasRole(roles...)
}

func (r *request) GetRoleId(role string) uint16 {
	roles := r.GetRoles()
	if roles == nil {
		return 0
	}
	return roles[role]
}

func (r *request) RequestUUID() string {
	return r.requestUUID
}

func (r *request) SetRequestUUID(str string) {
	r.requestUUID = str
}

func (r *request) SetKey(key string, value any) {
	r.tempMtx.Lock()
	defer r.tempMtx.Unlock()
	r.temp[key] = value
}

func (r *request) GetKey(key string) (value any, exists bool) {
	r.tempMtx.Lock()
	defer r.tempMtx.Unlock()
	v, ok := r.temp[key]
	return v, ok
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
	if err := r.context.Blob(getStatusCode(status), contentType, body); err != nil {
		return errors.HandleError(err)
	}
	return nil
}

func (r *request) RespondStream(status gateway.Status, contentType string, reader io.Reader) errors.ErrorModel {
	if err := r.context.Stream(getStatusCode(status), contentType, reader); err != nil {
		return errors.HandleError(err)
	}
	return nil
}

func (r *request) RespondFile(file string) errors.ErrorModel {
	if err := r.context.File(file); err != nil {
		return errors.HandleError(err)
	}
	return nil
}

func (r *request) RespondFsFile(file string, filesystem fs.FS) errors.ErrorModel {
	if err := r.context.File(file); err != nil {
		return errors.HandleError(err)
	}
	return nil
}

func (r *request) RespondHtml(status int, name string, body any) errors.ErrorModel {
	if err := r.context.Render(status, name, body); err != nil {
		return errors.HandleError(err)
	}
	return nil
}