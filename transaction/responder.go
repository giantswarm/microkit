// Package transaction provides trasnactional primitives to ensure certain
// actions happen only ones.
package transaction

import (
	"golang.org/x/net/context"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	microstorage "github.com/giantswarm/microkit/storage"
	transactionid "github.com/giantswarm/microkit/transaction/context/id"
)

// ResponderConfig represents the configuration used to create a responder.
type ResponderConfig struct {
	// Dependencies.
	Logger  micrologger.Logger
	Storage microstorage.Service
}

// DefaultResponderConfig provides a default configuration to create a new
// responder by best effort.
func DefaultResponderConfig() ResponderConfig {
	var err error

	var loggerService micrologger.Logger
	{
		loggerConfig := micrologger.DefaultConfig()
		loggerService, err = micrologger.New(loggerConfig)
		if err != nil {
			panic(err)
		}
	}

	var storageService microstorage.Service
	{
		storageConfig := microstorage.DefaultConfig()
		storageService, err = microstorage.New(storageConfig)
		if err != nil {
			panic(err)
		}
	}

	config := ResponderConfig{
		// Dependencies.
		Logger:  loggerService,
		Storage: storageService,
	}

	return config
}

// NewResponder creates a new configured responder.
func NewResponder(config ResponderConfig) (Responder, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}
	if config.Storage == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "storage must not be empty")
	}

	newResponder := &responder{
		// Dependencies.
		logger:  config.Logger,
		storage: config.Storage,
	}

	return newResponder, nil
}

type responder struct {
	// Dependencies.
	logger  micrologger.Logger
	storage microstorage.Service
}

func (r *responder) Reply(ctx context.Context, rr ResponseReplier) error {
	transactionID, ok := transactionid.FromContext(ctx)
	if !ok {
		return nil
	}
	transactionResponse, err := rr.responseStorage.Search(ctx, transactionID)
	if responsestorage.IsNotFound(err) {
		// There is no transaction for this request. The search was not successful,
		// which means the endpoint can be executed to its fullest extent.
		return nil
	} else if err != nil {
		return microerror.MaskAny(err)
	}

	// We found an existing transaction response. Here we write the tracked HTTP
	// headers.
	for key, val := range transactionResponse.Header {
		for _, h := range val {
			rr.Header().Add(key, h)
		}
	}

	rr.WriteHeader(transactionResponse.Code)

	_, err = rr.Write(transactionResponse.Body)
	if err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (r *responder) Track(ctx context.Context, rt ResponseTracker) error {
	transactionID, ok := transactionid.FromContext(ctx)
	if !ok {
		return nil
	}
	transactionResponse := responsestorage.Response{
		Body:   rt.BodyBuffer(),
		Code:   rt.StatusCode(),
		Header: rt.Header(),
	}

	err := r.responseStorage.Create(ctx, transactionID, transactionResponse)
	if err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}
