package main

import (
	"log"
	"net/http"
	"time"
)

type LogResponseWriter struct {
	ResponseWriter http.ResponseWriter
	statusCode     int
	size           int
}

func (lrw *LogResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LogResponseWriter) Write(b []byte) {
	if lrw.statusCode == 0 {
		lrw.statusCode = http.StatusOK
	}

	size, err = lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP()
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
