package server

import (
	"net/http"

	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

// Endpoint represents the management of transport logic. An endpoint defines
// what it needs to work properly. Internally it holds a reference to the
// service object which implements business logic and executes any workload.
// That means that network transport and business logic are strictly separated
// and work hand in hand via well defined package APIs.
type Endpoint interface {
	// Decoder returns the kithttp.DecodeRequestFunc used to decode a request
	// before the actual endpoint is executed.
	Decoder() kithttp.DecodeRequestFunc
	// Decoder returns the kithttp.EncodeResponseFunc used to encode a response
	// after the actual endpoint was executed.
	Encoder() kithttp.EncodeResponseFunc
	// Endpoint returns the kitendpoint.Endpoint which receives a decoded response
	// and forwards any workload to the internal service object reference.
	Endpoint() kitendpoint.Endpoint
	// Method returns the HTTP verb used to register the endpoint.
	Method() string
	// Middlewares returns the middlewares the endpoint configures to be run
	// before the actual endpoint is being invoked.
	Middlewares() []kitendpoint.Middleware
	// Name returns the name of the endpoint which can be used to label metrics or
	// annotate logs.
	Name() string
	// Path returns the HTTP request URL path used to register the endpoint.
	Path() string
}

// Server manages the HTTP transport logic.
type Server interface {
	// Boot registers the configured endpoints and starts the server under the
	// configured address.
	Boot()
	// Endpoints returns the server's configured list of endpoints. These are the
	// custom endpoints configured by the client.
	Endpoints() []Endpoint
	// ErrorEncoder returns the server's error encoder. This wraps the error
	// encoder configured by the client. Clients should not implement error
	// logging in here them self. This is done by the server itself. Clients must
	// not implement error response writing them self. This is done by the server
	// itself. Duplicated response writing will lead to runtime panics.
	ErrorEncoder() kithttp.ErrorEncoder
	// RequestFuncs returns the server's configured list of request functions.
	// These are the custom request functions configured by the client.
	RequestFuncs() []kithttp.RequestFunc
	// Shutdown stops the running server gracefully.
	Shutdown()
}

// ResponseWriter is a wrapper for http.ResponseWriter to track the written
// status code.
type ResponseWriter interface {
	// Header is only a wrapper around http.ResponseWriter.Header.
	Header() http.Header
	// StatusCode returns either the default status code of the one that was
	// actually written using WriteHeader.
	StatusCode() int
	// Write is only a wrapper around http.ResponseWriter.Write.
	Write(b []byte) (int, error)
	// WriteHeader is a wrapper around http.ResponseWriter.Write. In addition to
	// that it is used to track the written status code.
	WriteHeader(c int)
}
