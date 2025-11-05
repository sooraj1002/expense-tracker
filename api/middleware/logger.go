package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sooraj1002/expense-tracker/logger"
)

const maxLoggedBodySize = 2048

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	_, _ = w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	_, _ = w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// LoggerMiddleware logs HTTP requests and responses
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		var requestBody string
		requestHeaders := sanitizeHeaders(c.Request.Header.Clone())
		queryParams := formatQueryParams(c.Request.URL.Query())

		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Log.Errorw("Failed to read request body",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"error", err,
				)
			} else {
				requestBody = truncateBody(bodyBytes)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		logWriter := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = logWriter

		logger.Log.Infow("Incoming HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", queryParams,
			"ip", c.ClientIP(),
			"headers", requestHeaders,
			"body", requestBody,
		)

		// Process request
		c.Next()

		duration := time.Since(startTime)
		responseBody := truncateBody(logWriter.body.Bytes())
		responseHeaders := sanitizeHeaders(c.Writer.Header().Clone())

		logger.Log.Infow("Outgoing HTTP response",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration.String(),
			"headers", responseHeaders,
			"body", responseBody,
		)
	}
}

func truncateBody(body []byte) string {
	if len(body) <= maxLoggedBodySize {
		return string(body)
	}
	return string(body[:maxLoggedBodySize]) + "...(truncated)"
}

func sanitizeHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	sanitized := make(map[string]string, len(headers))
	for key, values := range headers {
		fmt.Printf("Header: %s => Values: %v\n", key, values)
		if shouldRedactHeader(key) {
			sanitized[key] = "[REDACTED]"
			continue
		}
		sanitized[key] = strings.Join(values, ",")
	}
	return sanitized
}

func shouldRedactHeader(key string) bool {
	switch strings.ToLower(key) {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key":
		return true
	default:
		return false
	}
}

func formatQueryParams(values map[string][]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	params := make(map[string]string, len(values))
	for key, vals := range values {
		params[key] = strings.Join(vals, ",")
	}
	return params
}
