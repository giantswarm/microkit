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
	"github.com/giantswarm/microkit/transaction"
	transactionid "github.com/giantswarm/microkit/transaction/context/id"
	transactiontracked "github.com/giantswarm/microkit/transaction/context/tracked"
)

// Config represents the configuration used to create a new server object.
type Config struct {
	// Dependencies.
	Endpoints            []Endpoint
	ErrorEncoder         kithttp.ErrorEncoder
	Logger               logger.Logger
	RequestFuncs         []kithttp.RequestFunc
	Router               *mux.Router
	TransactionResponder transaction.Responder

	// Settings.
	ListenAddress string
	ServiceName   string
}

// DefaultConfig provides a default configuration to create a new server object
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Endpoints:            nil,
		ErrorEncoder:         nil,
		Logger:               nil,
		RequestFuncs:         nil,
		Router:               mux.NewRouter(),
		TransactionResponder: nil,

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
	if config.TransactionResponder == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "transaction responder must not be empty")
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
		responder:    config.TransactionResponder,
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
	responder    transaction.Responder
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
				"code":  CodeResourceNotFound,
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
			errorEncoder := s.ErrorEncoder()
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
				kithttp.ServerErrorEncoder(errorEncoder),
			}

			// Register all endpoints to the router depending on their HTTP methods and
			// request paths. The registered http.Handler is instrumented using
			// prometheus. We track counts of execution and duration it took to complete
			// the http.Handler.
			s.router.Methods(method).Path(path).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var err error

				// Here we put the HTTP X-Transaction-ID header into the request
				// context, if any. We also check if there is a response already tracked
				// for the given transaction ID. This information is then stored within
				// the given request context as well.
				ctx := context.Background()
				{
					transactionID := r.Header.Get("X-Transaction-ID")
					if transactionID != "" {
						ctx = transactionid.NewContext(ctx, transactionID)

						exists, err := s.responder.Exists(ctx, transactionID)
						if err != nil {
							errorEncoder(ctx, err, w)
							return
						}
						ctx = transactiontracked.NewContext(ctx, exists)
					}
				}

				// Here we create a new wrapper for the http.ResponseWriter of the current
				// request. We inject it into the called http.Handler so it can track the
				// status code we are interested in. It will help us gathering the
				// response status code after it was written by the underlying
				// http.ResponseWriter.
				var responseWriter ResponseWriter
				{
					responseConfig := DefaultResponseWriterConfig()
					responseConfig.ResponseWriter = w
					responseWriter, err = NewResponseWriter(responseConfig)
					if err != nil {
						errorEncoder(ctx, err, w)
						return
					}
				}

				// TODO
				// Here we define the metrics relevant labels. These will be used to
				// instrument the current request.
				// This defered callback will be executed at the very end of the request.
				// When it is executed we know all necessary information to instrument the
				// complete request, including its response status code.
				{
					defer func(t time.Time) {
						endpointCode := strconv.Itoa(responseWriter.StatusCode())
						endpointMethod := strings.ToLower(method)
						endpointName := strings.Replace(name, "/", "_", -1)

						s.logger.Log("code", endpointCode, "endpoint", name, "method", endpointMethod, "path", r.URL.Path)

						endpointTotal.WithLabelValues(endpointCode, endpointMethod, endpointName).Inc()
						endpointTime.WithLabelValues(endpointCode, endpointMethod, endpointName).Set(float64(time.Since(t) / time.Millisecond))
					}(time.Now())
				}

				// TODO Here we know there does a transaction ID exist. We reply to the
				// request using the stored transaction response and return.
				wrappedDecoder := func(ctx context.Context, r *http.Request) (interface{}, error) {
					tracked, ok := transactiontracked.FromContext(ctx)
					if !ok {
						return nil, microerror.MaskAnyf(invalidContextError, "tracked must not be empty")
					}
					if tracked {
						transactionID, ok := transactionid.FromContext(ctx)
						if !ok {
							return nil, microerror.MaskAnyf(invalidContextError, "transaction ID must not be empty")
						}
						err := s.responder.Reply(ctx, transactionID, responseWriter)
						if err != nil {
							return nil, microerror.MaskAny(err)
						}

						return nil, nil
					}

					request, err := decoder(ctx, r)
					if err != nil {
						return nil, microerror.MaskAny(err)
					}

					return request, nil
				}

				// TODO
				wrappedEndpoint := func(ctx context.Context, request interface{}) (interface{}, error) {
					tracked, ok := transactiontracked.FromContext(ctx)
					if !ok {
						return nil, microerror.MaskAnyf(invalidContextError, "tracked must not be empty")
					}
					if tracked {
						return nil, nil
					}

					response, err := endpoint(ctx, request)
					if err != nil {
						return nil, microerror.MaskAny(err)
					}

					return response, nil
				}

				// TODO Here we know there does a transaction response is known. We
				// already replied to the request using the stored transaction
				// response within the wrapped decoder. Thus we do not have to do
				// anything here.
				wrappedEncoder := func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
					tracked, ok := transactiontracked.FromContext(ctx)
					if !ok {
						return microerror.MaskAnyf(invalidContextError, "tracked must not be empty")
					}
					if tracked {
						return nil
					}

					err := encoder(ctx, w, response)
					if err != nil {
						return microerror.MaskAny(err)
					}

					transactionID, ok := transactionid.FromContext(ctx)
					if !ok {
						// In case the response is not already tracked, but there is no
						// transaction ID, we cannot track it at all. So we return here.
						return nil
					}
					err = s.responder.Track(ctx, transactionID, responseWriter)
					if err != nil {
						return microerror.MaskAny(err)
					}

					return nil
				}

				// Now we execute the actual go-kit endpoint handler.
				kithttp.NewServer(
					ctx,
					wrappedEndpoint,
					wrappedDecoder,
					wrappedEncoder,
					options...,
				).ServeHTTP(responseWriter, r)
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
	return func(ctx context.Context, serverError error, w http.ResponseWriter) {
		var err error

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
			responseConfig.Underlying = serverError
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
		errDomain := errorDomain(serverError)
		errMessage := errorMessage(serverError)
		errTrace := errorTrace(serverError)
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
