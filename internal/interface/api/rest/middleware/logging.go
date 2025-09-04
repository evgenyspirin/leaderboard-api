package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const maxLogBodySize = 1 << 12 // 4 KB

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (ww *wrappedWriter) WriteHeader(code int) {
	ww.statusCode = code
	ww.ResponseWriter.WriteHeader(code)
}

func RequestLog(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// todo: debug level(dev/prod) / mask sensitive event data
			var body string
			if r.Body != nil {
				var buf bytes.Buffer
				limited := io.LimitReader(r.Body, maxLogBodySize)
				io.Copy(&buf, limited)
				body = buf.String()
				r.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))
			}

			ww := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(ww, r)

			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.Int("status", ww.statusCode),
				zap.Duration("duration", time.Since(start)),
				zap.String("body", body),
			)
		})
	}
}
