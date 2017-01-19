// Package server provides a server implementation to connect network transport
// protocols and service business logic by defining server endpoints.
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tylerb/graceful"
	"golang.org/x/net/context"

	microerror "github.com/giantswarm/microkit/error"
	"github.com/giantswarm/microkit/logger"
)

// Config represents the configuration used to create a new server object.
type Config struct {
	// Dependencies.
	Endpoints    []Endpoint
	ErrorEncoder kithttp.ErrorEncoder
	Logger       logger.Logger
	RequestFuncs []kithttp.RequestFunc
	Router       *mux.Router

	// Settings.
	ListenAddress string
	ServiceName   string
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
		Router:       mux.NewRouter(),

		// Settings.
		ListenAddress: "http://127.0.0.1:8000",
		ServiceName:   "microkit",
	}
}

// New creates a new configured server object.
func New(config Config) (Server, error) {
	// Dependencies.
	if config.Endpoints == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "endpoints must not be empty")
	}
	if config.ErrorEncoder == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "error encoder must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}
	if config.RequestFuncs == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "request funcs must not be empty")
	}
	if config.Router == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "router must not be empty")
	}

	// Settings.
	if config.ListenAddress == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "listen address must not be empty")
	}
	if config.ServiceName == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "service name must not be empty")
	}

	listenURL, err := url.Parse(config.ListenAddress)
	if err != nil {
		return nil, microerror.MaskAnyf(invalidConfigError, err.Error())
	}

	newServer := &server{
		bootOnce:     sync.Once{},
		endpoints:    config.Endpoints,
		errorEncoder: config.ErrorEncoder,
		httpServer:   nil,
		listenURL:    listenURL,
		logger:       config.Logger,
		requestFuncs: config.RequestFuncs,
		router:       config.Router,
		serviceName:  config.ServiceName,
		shutdownOnce: sync.Once{},
	}

	return newServer, nil
}

// server manages the transport logic and endpoint registration.
type server struct {
	bootOnce     sync.Once
	endpoints    []Endpoint
	errorEncoder kithttp.ErrorEncoder
	httpServer   *graceful.Server
	listenURL    *url.URL
	logger       logger.Logger
	requestFuncs []kithttp.RequestFunc
	router       *mux.Router
	serviceName  string
	shutdownOnce sync.Once
}

func (s *server) Boot() {
	s.bootOnce.Do(func() {
		// Define our custom not found handler. Here we take care about logging,
		// metrics and a proper response.
		s.router.NotFoundHandler = http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log the error and its message. This is really useful for debugging.
			errDomain := errorDomain(nil)
			errMessage := fmt.Sprintf("not found: %s %s", r.Method, r.URL.Path)
			errTrace := ""
			s.logger.Log("error", map[string]string{"domain": errDomain, "message": errMessage, "trace": errTrace})

			// This defered callback will be executed at the very end of the request.
			defer func(t time.Time) {
				endpointCode := strconv.Itoa(http.StatusNotFound)
				endpointMethod := strings.ToLower(r.Method)
				endpointName := "notfound"

				endpointTotal.WithLabelValues(endpointCode, endpointMethod, endpointName).Inc()
				endpointTime.WithLabelValues(endpointCode, endpointMethod, endpointName).Set(float64(time.Since(t) / time.Millisecond))

				errorTotal.WithLabelValues(errDomain).Inc()
			}(time.Now())

			// Write the actual response body.
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": errMessage,
				"from":  s.ServiceName(),
			})
		}))

		// We go through all endpoints this server defines and register them to the
		// router.
		for _, e := range s.Endpoints() {
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
			s.router.Methods(method).Path(path).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.Background()

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
					s.logger.Log("code", endpointCode, "endpoint", name, "method", endpointMethod, "path", r.URL.Path)

					// At the time this code is executed the status code is properly set. So
					// we can use it for our instrumentation.
					endpointCode := strconv.Itoa(endpointCode)
					endpointTotal.WithLabelValues(endpointCode, endpointMethod, endpointName).Inc()
					endpointTime.WithLabelValues(endpointCode, endpointMethod, endpointName).Set(float64(time.Since(t) / time.Millisecond))
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

		// Register prometheus metrics endpoint.
		s.router.Path("/metrics").Handler(promhttp.Handler())

		// Register the router which has all of the configured custom endpoints
		// registered.
		s.httpServer = &graceful.Server{
			NoSignalHandling: true,
			Server: &http.Server{
				Addr:    s.listenURL.Host,
				Handler: s.router,
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
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		// At first we have to set the content type of the actual error response. If
		// we would set it at the end we would set a trailing header that would not
		// be recognized by most of the clients out there. This is because in the
		// next call to the errorEncoder below the client's implementation of the
		// errorEncoder probably writes the status code header, which marks the
		// beginning of trailing headers in HTTP.
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		// Create the microkit specific response error, which acts as error wrapper
		// within the client's error encoder. It is used to propagate response codes
		// and messages, so we can use them below.
		var responseError ResponseError
		{
			responseConfig := DefaultResponseErrorConfig()
			responseConfig.Underlying = err
			responseError, err = NewResponseError(responseConfig)
			if err != nil {
				panic(err)
			}
		}

		// Run the custom error encoder. This is used to let the implementing
		// microservice do something with errors occured during runtime. Things like
		// writing specific HTTP status codes to the given response writer can be
		// done.
		s.errorEncoder(ctx, responseError, w)

		// Log the error and its errgo trace. This is really useful for debugging.
		errDomain := errorDomain(err)
		errMessage := errorMessage(err)
		errTrace := errorTrace(err)
		s.logger.Log("error", map[string]string{"domain": errDomain, "message": errMessage, "trace": errTrace})

		// Emit metrics about the occured errors. That way we can feed our
		// instrumentation stack to have nice dashboards to get a picture about the
		// general system health.
		errorTotal.WithLabelValues(errDomain).Inc()

		// Write the actual response body.
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":  responseError.Code(),
			"error": responseError.Message(),
			"from":  s.ServiceName(),
		})
	}
}

func (s *server) RequestFuncs() []kithttp.RequestFunc {
	return s.requestFuncs
}

// NewRouter returns a HTTP handler for the server. Here we register all
// endpoints listed in the endpoint collection.
func (s *server) Router() *mux.Router {
	return s.router
}

func (s *server) ServiceName() string {
	return s.serviceName
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
