package server

import (
	"net/http"
)

// ResponseWriterConfig represents the configuration used to create a new
// response writer.
type ResponseWriterConfig struct {
	// Settings.
	ResponseWriter http.ResponseWriter
	StatusCode     int
}

// DefaultResponseWriterConfig provides a default configuration to create a new
// response writer by best effort.
func DefaultResponseWriterConfig() ResponseWriterConfig {
	return ResponseWriterConfig{
		// Settings.
		ResponseWriter: nil,
		StatusCode:     http.StatusOK,
	}
}

// New creates a new configured response writer.
func NewResponseWriter(config ResponseWriterConfig) (ResponseWriter, error) {
	// Settings.
	if config.ResponseWriter == nil {
		return nil, maskAnyf(invalidConfigError, "response writer must not be empty")
	}
	if config.StatusCode == 0 {
		return nil, maskAnyf(invalidConfigError, "status code must not be empty")
	}

	newResponseWriter := &responseWriter{
		responseWriter: config.ResponseWriter,
		statusCode:     config.StatusCode,
	}

	return newResponseWriter, nil
}

type responseWriter struct {
	responseWriter http.ResponseWriter
	statusCode     int
}

func (rw *responseWriter) Header() http.Header {
	return rw.responseWriter.Header()
}

func (rw *responseWriter) StatusCode() int {
	return rw.statusCode
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.responseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(c int) {
	rw.responseWriter.WriteHeader(c)
	rw.statusCode = c
}
