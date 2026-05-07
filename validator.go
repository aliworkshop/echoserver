package echoserver

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

const validatorContextKey = "_echoserver.validator"

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func injectValidator(v *validator.Validate) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(validatorContextKey, v)
			return next(c)
		}
	}
}

func getValidator(c echo.Context) *validator.Validate {
	if v, ok := c.Get(validatorContextKey).(*validator.Validate); ok {
		return v
	}
	return validator.New()
}
