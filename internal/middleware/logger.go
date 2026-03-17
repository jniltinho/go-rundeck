package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// RequestLogger returns a middleware that logs each HTTP request.
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			req := c.Request()
			res := c.Response()
			elapsed := time.Since(start)

			c.Logger().Infof(
				"method=%s path=%s status=%d latency=%s ip=%s",
				req.Method,
				req.URL.Path,
				res.Status,
				elapsed,
				c.RealIP(),
			)
			_ = log.INFO // keep import used
			return err
		}
	}
}
