// Package server provides a server implementation to connect network transport
// protocols and service business logic by defining server endpoints.
package server

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	kitendpoint "github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tylerb/graceful"
	"golang.org/x/net/context"

	"github.com/giantswarm/microkit/logger"
)

// Config represents the configuration used to create a new server object.
type Config struct {
	// Dependencies.
	Endpoints    []Endpoint
	ErrorEncoder kithttp.ErrorEncoder
	Logger       logger.Logger
	RequestFuncs []kithttp.RequestFunc

	// Settings.
	ListenAddress string
}

// DefaultConfig provides a default configuration to create a new server object
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Endpoints:    nil,
		ErrorEncoder: nil,
		Logger:       nil,
		RequestFuncs: nil,

		// Settings.
		ListenAddress: "http://127.0.0.1:8080",
	}
}

// New creates a new configured server object.
func New(config Config) (Server, error) {
	// Dependencies.
	if config.Endpoints == nil {
		return nil, maskAnyf(invalidConfigError, "endpoints must not be empty")
	}
	if config.ErrorEncoder == nil {
		return nil, maskAnyf(invalidConfigError, "error encoder must not be empty")
	}
	if config.Logger == nil {
		return nil, maskAnyf(invalidConfigError, "logger must not be empty")
	}
	if config.RequestFuncs == nil {
		return nil, maskAnyf(invalidConfigError, "request funcs must not be empty")
	}

	// Settings.
	if config.ListenAddress == "" {
		return nil, maskAnyf(invalidConfigError, "listen address must not be empty")
	}

	listenURL, err := url.Parse(config.ListenAddress)
	if err != nil {
		return nil, maskAnyf(invalidConfigError, err.Error())
	}

	newServer := &server{
		Config: config,

		bootOnce:     sync.Once{},
		endpoints:    config.Endpoints,
		errorEncoder: config.ErrorEncoder,
		httpServer:   nil,
		listenURL:    listenURL,
		logger:       kitlog.NewContext(config.Logger).With("package", "server"),
		requestFuncs: config.RequestFuncs,
		shutdownOnce: sync.Once{},
	}

	return newServer, nil
}

// server manages the transport logic and endpoint registration.
type server struct {
	Config

	bootOnce     sync.Once
	endpoints    []Endpoint
	errorEncoder kithttp.ErrorEncoder
	httpServer   *graceful.Server
	listenURL    *url.URL
	logger       logger.Logger
	requestFuncs []kithttp.RequestFunc
	shutdownOnce sync.Once
}

// Boot registers the configured endpoints and starts the server under the
// configured address.
func (s *server) Boot() {
	s.bootOnce.Do(func() {
		handler := s.NewRouter()

		// Register prometheus metrics endpoint.
		handler.Path("/metrics").Handler(promhttp.Handler())

		// Register the router which has all of the configured custom endpoints
		// registered.
		s.httpServer = &graceful.Server{
			NoSignalHandling: true,
			Server: &http.Server{
				Addr:    s.listenURL.Host,
				Handler: handler,
			},
			Timeout: 3 * time.Second,
		}

		go func() {
			err := s.httpServer.ListenAndServe()
			if err != nil {
				panic(err)
			}
		}()
	})
}

func (s *server) Endpoints() []Endpoint {
	return s.endpoints
}

func (s *server) ErrorEncoder() kithttp.ErrorEncoder {
	return s.errorEncoder
}

// NewRouter returns a HTTP handler for the server. Here we register all
// endpoints listed in the endpoint collection.
func (s *server) NewRouter() *mux.Router {
	router := mux.NewRouter()

	// We go through all endpoints this server defines and register them to the
	// router.
	for _, e := range s.Endpoints() {
		ctx := context.Background()

		decoder := e.Decoder()
		encoder := e.Encoder()
		endpoint := e.Endpoint()
		method := e.Method()
		middlewares := e.Middlewares()
		name := e.Name()
		path := e.Path()

		// Prepare the actual endpoint depending on the provided middlewares of the
		// endpoint implementation. There might be cases in which there are none or
		// only one middleware. The go-kit interface is not that nice so we need to
		// make it fit here.
		{
			if len(middlewares) == 1 {
				endpoint = kitendpoint.Chain(middlewares[0])(endpoint)
			}
			if len(middlewares) > 1 {
				endpoint = kitendpoint.Chain(middlewares[0], middlewares[1:]...)(endpoint)
			}
		}

		// Combine all options this server defines.
		options := []kithttp.ServerOption{
			kithttp.ServerBefore(s.RequestFuncs()...),
			kithttp.ServerErrorEncoder(s.ErrorEncoder()),
		}

		// Register all endpoints to the router depending on their HTTP methods and
		// request paths. The registered http.Handler is instrumented using
		// prometheus. We track counts of execution and duration it took to complete
		// the http.Handler.
		router.Methods(method).Path(path).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Here we create a new wrapper for the http.ResponseWriter of the current
			// request. We inject it into the called http.Handler so it can track the
			// status code we are interested in. It will help us gathering the
			// response status code after it was written by the underlying
			// http.ResponseWriter.
			responseWriterConfig := DefaultResponseWriterConfig()
			responseWriterConfig.ResponseWriter = w
			responseWriter, err := NewResponseWriter(responseWriterConfig)
			if err != nil {
				panic(err)
			}
			w = responseWriter

			// Here we define the metrics relevant labels. These will be used to
			// instrument the current request.
			endpointName := strings.Replace(name, "/", "_", -1)
			endpointMethod := strings.ToLower(method)
			endpointCode := responseWriter.StatusCode()

			// This defered callback will be executed at the very end of the request.
			// When it is executed we know all necessary information to instrument the
			// complete request, including its response status code.
			defer func(t time.Time) {
				// At the time this code is executed the status code is properly set. So
				// we can use it for our instrumentation.
				endpointCode := strconv.Itoa(endpointCode)
				endpointTotal.WithLabelValues(endpointMethod, endpointName, endpointCode).Inc()
				endpointTime.WithLabelValues(endpointMethod, endpointName, endpointCode).Set(float64(time.Since(t) / time.Millisecond))
			}(time.Now())

			// Now we execute the actual endpoint handler.
			kithttp.NewServer(
				ctx,
				endpoint,
				decoder,
				encoder,
				options...,
			).ServeHTTP(w, r)

			// Here we now the status code.
			endpointCode = responseWriter.StatusCode()
		}))
	}

	return router
}

func (s *server) RequestFuncs() []kithttp.RequestFunc {
	return s.requestFuncs
}

func (s *server) Shutdown() {
	s.shutdownOnce.Do(func() {
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			// Stop the HTTP server gracefully and wait some time for open connections
			// to be closed. Then force it to be stopped.
			s.httpServer.Stop(s.httpServer.Timeout)
			<-s.httpServer.StopChan()
			wg.Done()
		}()

		wg.Wait()
	})
}
