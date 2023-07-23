package echoserver

import (
	"github.com/aliworkshop/error"
	"github.com/aliworkshop/gateway/v2"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"net/http"
)

func NewResponder(languageBundle *i18n.Bundle) gateway.Responder {
	return &echoResponder{
		languageBundle: languageBundle,
	}
}

func NewEmptyResponder(languageBundle *i18n.Bundle) gateway.Responder {
	return &emptyResponder{
		languageBundle: languageBundle,
	}
}

type echoResponder struct {
	languageBundle *i18n.Bundle
	total          uint
}
type emptyResponder struct {
	languageBundle *i18n.Bundle
	total          uint
}

func (er *echoResponder) LanguageBundle() *i18n.Bundle {
	return er.languageBundle
}

func (er *echoResponder) SetLanguageBundle(bundle *i18n.Bundle) {
	er.languageBundle = bundle
}

func (er *echoResponder) SetTotal(total uint) gateway.Responder {
	er.total = total
	return er
}

func (er *echoResponder) Respond(req gateway.Requester, status gateway.Status, result interface{}) {
	ctx := req.GetContext().(echo.Context)
	req.SetResponded(true)
	ctx.Request().Header.Set("X-Request-UID", req.GetUid())
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
	case gateway.StatusUnknown:
		ctx = req.GetContext().(echo.Context)
		switch ctx.Request().Method {
		case "POST", "PUT":
			if result == nil {
				status = gateway.StatusNoContent
				break
			}
			status = gateway.StatusCreated
		case "GET":
			status = gateway.StatusOK
		case "DELETE":
			if result == nil {
				status = gateway.StatusNoContent
				break
			}
			status = gateway.StatusOK
		}
		break
	}
	response := gateway.Response{
		Page:    req.Paginator().Page(),
		PerPage: req.Paginator().PerPage(),
		Items:   result,
		Total:   req.Paginator().Total(),
	}
	ctx.JSON(getStatusCode(status), response)
}

func (er *echoResponder) RespondError(req gateway.Requester, err error.ErrorModel) {
	ctx := req.GetContext().(echo.Context)
	ctx.Request().Header.Set("X-Request-UID", req.GetUid())
	if er.languageBundle != nil {
		errId := err.Id()
		if errId != "" && (err.IsMsgDefault() || !err.IsIdDefault()) {
			err = err.Clone().WithMessage(req.ShouldLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    errId,
					Other: err.Message(),
					One:   err.MessageOne(),
				},
				TemplateData: err.Properties(),
				PluralCount:  err.PluralCount(),
			}))
		}
	}
	ctx.JSON(getStatusCodeByError(err), err)
	req.SetResponded(true)
}

func (er *emptyResponder) LanguageBundle() *i18n.Bundle {
	return er.languageBundle
}

func (er *emptyResponder) SetLanguageBundle(bundle *i18n.Bundle) {
	er.languageBundle = bundle
}

func (er *emptyResponder) SetTotal(total uint) gateway.Responder {
	er.total = total
	return er
}

func (er *emptyResponder) Respond(req gateway.Requester, status gateway.Status, result interface{}) {
	ctx := req.GetContext().(echo.Context)
	ctx.Request().Header.Set("X-Request-UID", req.GetUid())
	ctx.JSON(getStatusCode(status), result)
	req.SetResponded(true)
}

func (er *emptyResponder) RespondError(req gateway.Requester, err error.ErrorModel) {
	ctx := req.GetContext().(echo.Context)
	ctx.Request().Header.Set("X-Request-UID", req.GetUid())
	ctx.Error(err)
	req.SetResponded(true)
}
