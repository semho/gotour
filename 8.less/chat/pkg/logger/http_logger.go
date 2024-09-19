package logger

import (
	"io"
	"net/http"
	"strings"
)

type captureResponseWriter struct {
	http.ResponseWriter
	statusCode int
	response   []byte
}

func (w *captureResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *captureResponseWriter) Write(b []byte) (int, error) {
	w.response = append(w.response, b...)
	return w.ResponseWriter.Write(b)
}

func LogRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			Log.Info(
				"Incoming request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Read and log request body
			body, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(strings.NewReader(string(body)))
			Log.Info("Request body", "body", string(body))

			// Capture the response
			crw := &captureResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			handler.ServeHTTP(crw, r)

			Log.Info(
				"Response",
				"status_code", crw.statusCode,
				"body", string(crw.response),
			)
		},
	)
}
