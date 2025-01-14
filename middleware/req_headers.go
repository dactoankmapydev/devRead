package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"mime"
	"net/http"

	"devread/model"
)

func HeadersMiddleware() echo.MiddlewareFunc {
	config := middleware.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		//ReferrerPolicy:        "origin",
		//ContentSecurityPolicy: "default-src 'self'",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: true,
		HSTSPreloadEnabled:    true,
	}
	return middleware.SecureWithConfig(config)
}

func HeadersAccept() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			contentType := c.Request().Header.Get("Content-Type")

			if contentType != "" {
				mt, _, err := mime.ParseMediaType(contentType)
				if err != nil {
					return c.JSON(http.StatusBadRequest, model.Response{
						StatusCode: http.StatusBadRequest,
						Message:    err.Error(),
					})
				}

				if mt != "application/json" {
					return c.JSON(http.StatusUnsupportedMediaType, model.Response{
						StatusCode: http.StatusUnsupportedMediaType,
						Message:    err.Error(),
					})
				}
			}

			return next(c)
		}
	}
}

func CORSMiddleware() echo.MiddlewareFunc {
	config := middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderContentLength, echo.HeaderAccept, echo.HeaderOrigin},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}
	return middleware.CORSWithConfig(config)
}
