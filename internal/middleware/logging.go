package middleware

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type responseData struct {
	status int
	size   int
}
type responseWrite struct {
	http.ResponseWriter
	responseData *responseData
}

func (rw *responseWrite) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseData.size += size
	return size, err
}

func (rw *responseWrite) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.responseData.status = statusCode
}

func Logging(logger *zap.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ts := time.Now()
			responseData := &responseData{
				status: http.StatusOK,
				size:   0,
			}

			lrw := responseWrite{
				ResponseWriter: w,
				responseData:   responseData,
			}

			next.ServeHTTP(&lrw, r)

			dt := time.Since(ts)

			logger.Info("got incoming HTTP request",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("status", responseData.status),
				zap.Int("size", responseData.size),
				zap.Duration("duration", dt),
			)
		})
	}
}
