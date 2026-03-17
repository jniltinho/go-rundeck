package middleware

import (
	"time"

	"github.com/labstack/echo/v5"
	"log/slog"
	"net/http"
)

// RequestLogger returns a middleware that logs each HTTP request.
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)
			req := c.Request()
			res := c.Response()
			elapsed := time.Since(start)

			status := http.StatusOK
			if echoRes, unwrapErr := echo.UnwrapResponse(res); unwrapErr == nil {
				status = echoRes.Status
			}

			c.Logger().Info(
				"request completed",
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.Int("status", status),
				slog.Duration("latency", elapsed),
				slog.String("ip", c.RealIP()),
			)
			return err
		}
	}
}
