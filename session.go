package echoserver

import (
	"fmt"
	"github.com/aliworkshop/handlerlib/authorization"
	"github.com/labstack/echo/v4"
	"net/url"
)

func (rh *router) getCsrfConfig(sessionType string) *CSRFConfig {
	csrfConf := rh.config.CSRF.SessionTypes[sessionType]
	if csrfConf == nil {
		csrfConf = rh.config.CSRF.SessionTypes["DEFAULT"]
	}
	return csrfConf
}

func (rh *router) verifySession(c echo.Context, session authorization.SessionModel) error {
	csrfToken := session.GetCSRFToken()
	if csrfToken != "" {
		csrfConf := rh.getCsrfConfig(session.GetType())
		if csrfConf == nil {
			return fmt.Errorf("csrf config not found")
		}
		// handle CSRF token
		csrfTokenHeader := c.Request().Header.Get(csrfConf.HeaderKey)
		csrfTokenCookie, _ := c.Cookie(csrfConf.CookieKey)
		val, _ := url.QueryUnescape(csrfTokenCookie.Value)
		if csrfToken != csrfTokenHeader || csrfToken != val {
			return fmt.Errorf("csrf does not match")
		}
	}
	return nil
}
