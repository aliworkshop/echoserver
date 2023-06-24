package echoserver

import (
	"github.com/aliworkshop/errorslib"
	"github.com/aliworkshop/handlerlib"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"net/http"
)

func NewResponder(languageBundle *i18n.Bundle) handlerlib.Responder {
	return &echoResponder{
		languageBundle: languageBundle,
	}
}

func NewEmptyResponder(languageBundle *i18n.Bundle) handlerlib.Responder {
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

func (er *echoResponder) SetTotal(total uint) handlerlib.Responder {
	er.total = total
	return er
}

func (er *echoResponder) Respond(req handlerlib.RequestModel, status handlerlib.Status, result interface{}) {
	if f, ok := result.(handlerlib.ResponseFinalizer); ok {
		f.Finalize()
	}
	ctx := req.GetContext().(echo.Context)
	req.SetResponded(true)
	ctx.Request().Header.Set("X-Request-UID", req.GetUid())
	switch status {
	case handlerlib.StatusMovedPermanently:
		ctx.Redirect(http.StatusMovedPermanently, result.(string))
		return
	case handlerlib.StatusFound:
		ctx.Redirect(http.StatusFound, result.(string))
		return
	case handlerlib.StatusPermanentRedirect:
		ctx.Redirect(http.StatusPermanentRedirect, result.(string))
		return
	case handlerlib.StatusTemporaryRedirect:
		ctx.Redirect(http.StatusTemporaryRedirect, result.(string))
		return
	case "":
		ctx = req.GetContext().(echo.Context)
		switch ctx.Request().Method {
		case "POST", "PUT":
			if result == nil {
				status = handlerlib.StatusNoContent
				break
			}
			status = handlerlib.StatusCreated
		case "GET":
			status = handlerlib.StatusOK
		case "DELETE":
			if result == nil {
				status = handlerlib.StatusNoContent
				break
			}
			status = handlerlib.StatusOK
		}
		break
	}
	response := handlerlib.Response{
		Page:    req.Paging().Page(),
		PerPage: req.Paging().PerPage(),
		Items:   result,
		Total:   er.total,
	}
	ctx.JSON(getStatusCode(status), response)
}

func (er *echoResponder) RespondWithError(req handlerlib.RequestModel, err errorslib.ErrorModel) {
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

func (er *emptyResponder) SetTotal(total uint) handlerlib.Responder {
	er.total = total
	return er
}

func (er *emptyResponder) Respond(req handlerlib.RequestModel, status handlerlib.Status, result interface{}) {
	ctx := req.GetContext().(echo.Context)
	ctx.Request().Header.Set("X-Request-UID", req.GetUid())
	ctx.JSON(getStatusCode(status), result)
	req.SetResponded(true)
}

func (er *emptyResponder) RespondWithError(req handlerlib.RequestModel, err errorslib.ErrorModel) {
	ctx := req.GetContext().(echo.Context)
	ctx.Request().Header.Set("X-Request-UID", req.GetUid())
	ctx.Error(err)
	req.SetResponded(true)
}
