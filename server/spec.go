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

type Server interface {
	Boot()
	Endpoints() []Endpoint
	ErrorEncoder() kithttp.ErrorEncoder
	RequestFuncs() []kithttp.RequestFunc
	Shutdown()
}

type ResponseWriter interface {
	Header() http.Header
	StatusCode() int
	Write(b []byte) (int, error)
	WriteHeader(c int)
}
