package main

import (
	"log"
	"net/http"
	"time"
)

type LogResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

var disableLogs bool

func (lrw *LogResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LogResponseWriter) Write(b []byte) (int, error) {
	if lrw.statusCode == 0 {
		lrw.statusCode = http.StatusOK
	}

	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

func InitializeLogResponseWriter(disable bool) {
	disableLogs = disable
}

func loggingMiddleware(next http.Handler) http.Handler {
	if disableLogs {
		return next
	}
	return requestLogger(next)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &LogResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		target := w.Header().Get("X-Forwarded-Server")
		log.Printf("[%s] %s %s -> server: %s | status: %d | TAT: %v | size: %dB",
			r.Method,
			r.RemoteAddr,
			r.URL.Path,
			target,
			lrw.statusCode,
			duration,
			lrw.size)
	})
}
