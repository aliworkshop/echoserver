package echoserver

import (
	"net/http"

	"github.com/aliworkshop/errors"
	"github.com/aliworkshop/gateway/v2"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func NewResponder(languageBundle *i18n.Bundle) gateway.Responder {
	return &echoResponder{languageBundle: languageBundle}
}

func NewEmptyResponder(languageBundle *i18n.Bundle) gateway.Responder {
	return &emptyResponder{languageBundle: languageBundle}
}

type echoResponder struct {
	languageBundle *i18n.Bundle
}

type emptyResponder struct {
	languageBundle *i18n.Bundle
}

func (er *echoResponder) LanguageBundle() *i18n.Bundle {
	return er.languageBundle
}

func (er *echoResponder) Respond(req gateway.HttpRequester, status gateway.Status, result any) {
	ctx := req.GetHttpContext().(echo.Context)
	ctx.Response().Header().Set("X-Request-Uuid", req.RequestUUID())
	req.SetIsResponded(true)

	if status == gateway.StatusUnknown {
		switch req.GetMethod() {
		case http.MethodPost, http.MethodPut:
			if result == nil {
				status = gateway.StatusNoContent
				break
			}
			status = gateway.StatusCreated
		case http.MethodGet:
			status = gateway.StatusOK
		case http.MethodDelete:
			if result == nil {
				status = gateway.StatusNoContent
				break
			}
			status = gateway.StatusOK
		default:
			status = gateway.StatusOK
		}
	}

	switch status {
	case gateway.StatusMovedPermanently:
		ctx.Redirect(http.StatusMovedPermanently, result.(string))
		return
	case gateway.StatusFound:
		ctx.Redirect(http.StatusFound, result.(string))
		return
	case gateway.StatusPermanentRedirect:
		ctx.Redirect(http.StatusPermanentRedirect, result.(string))
		return
	case gateway.StatusTemporaryRedirect:
		ctx.Redirect(http.StatusTemporaryRedirect, result.(string))
		return
	case gateway.StatusNoContent:
		ctx.NoContent(http.StatusNoContent)
		return
	}

	p := req.Paginator()
	response := gateway.Response{
		Page:    int(p.GetPage()),
		PerPage: int(p.GetPageSize()),
		Items:   result,
		Total:   p.Total(),
	}
	ctx.JSON(getStatusCode(status), response)
}

func (er *echoResponder) RespondError(req gateway.HttpRequester, err errors.ErrorModel) {
	ctx := req.GetHttpContext().(echo.Context)
	ctx.Response().Header().Set("X-Request-Uuid", req.RequestUUID())
	if er.languageBundle != nil {
		errId := err.Id()
		if errId != "" && (err.IsMsgDefault() || !err.IsIdDefault() || len(err.Properties()) > 0) {
			err = err.Clone().WithMessage(req.ShouldLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    errId,
					Other: err.Message(),
				},
				TemplateData: err.Properties(),
			}))
		}
	}
	ctx.JSON(getStatusCodeByError(err), err)
	req.SetIsResponded(true)
}

func (er *emptyResponder) LanguageBundle() *i18n.Bundle {
	return er.languageBundle
}

func (er *emptyResponder) Respond(req gateway.HttpRequester, status gateway.Status, result any) {
	ctx := req.GetHttpContext().(echo.Context)
	ctx.Response().Header().Set("X-Request-Uuid", req.RequestUUID())
	ctx.JSON(getStatusCode(status), result)
	req.SetIsResponded(true)
}

func (er *emptyResponder) RespondError(req gateway.HttpRequester, err errors.ErrorModel) {
	ctx := req.GetHttpContext().(echo.Context)
	ctx.Response().Header().Set("X-Request-Uuid", req.RequestUUID())
	ctx.JSON(getStatusCodeByError(err), err)
	req.SetIsResponded(true)
}
