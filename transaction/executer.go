// Package transaction provides transactional primitives to ensure certain
// actions happen only ones.
package transaction

import (
	"fmt"

	"golang.org/x/net/context"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	microstorage "github.com/giantswarm/microkit/storage"
	transactionid "github.com/giantswarm/microkit/transaction/context/id"
)

// ExecuterConfig represents the configuration used to create a executer.
type ExecuterConfig struct {
	// Dependencies.
	Logger  micrologger.Logger
	Storage microstorage.Service
}

// DefaultExecuterConfig provides a default configuration to create a new
// executer by best effort.
func DefaultExecuterConfig() ExecuterConfig {
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

	config := ExecuterConfig{
		// Dependencies.
		Logger:  loggerService,
		Storage: storageService,
	}

	return config
}

// NewExecuter creates a new configured executer.
func NewExecuter(config ExecuterConfig) (Executer, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}
	if config.Storage == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "storage must not be empty")
	}

	newExecuter := &executer{
		// Dependencies.
		logger:  config.Logger,
		storage: config.Storage,
	}

	return newExecuter, nil
}

type executer struct {
	// Dependencies.
	logger  micrologger.Logger
	storage microstorage.Service
}

func (e *executer) Execute(ctx context.Context, config ExecuteConfig) error {
	// Validate the execute config to make sure we can safely work with it.
	err := validateExecuteConfig(config)
	if err != nil {
		return microerror.MaskAny(err)
	}

	// At first we check for the transaction ID that might be obtained by the
	// given context. We actually do not care about if there is one or not. It
	// will be either set or empty. In case it is set, we use it for the execution
	// of the transaction below. In case it is not set at all, we simply want to
	// execute the configured trial all the time.
	transactionID, ok := transactionid.FromContext(ctx)
	if !ok {
		err := config.Trial(ctx)
		if err != nil {
			return microerror.MaskAny(err)
		}

		e.logger.Log("debug", fmt.Sprintf("executed transaction trial for transaction ID %s and trial ID %s", transactionID, config.TrialID))

		return nil
	}

	// Here we know there is a transaction ID given. Thus we want to check if the
	// trial was already successful. If the trial was already successful we want
	// to execute the configured replay, if any. If there was no trial for the
	// given transaction registered yet, we are executing the trial.
	{
		key := transactionKey("transaction", transactionID, "trial", config.TrialID)
		exists, err := e.storage.Exists(ctx, key)
		if err != nil {
			return microerror.MaskAny(err)
		}

		if exists && config.Replay != nil {
			err := config.Replay(ctx)
			if err != nil {
				return microerror.MaskAny(err)
			}

			e.logger.Log("debug", fmt.Sprintf("executed transaction replay for transaction ID %s and trial ID %s", transactionID, config.TrialID))

			return nil
		}
	}

	// Here we know we have to execute the trial. In case the trial failed we
	// simply return the error. In case the trial was successful we register it to
	// be sure we already executed it. That causes the trial to be ignored the
	// next time the transaction is being executed and the transaction's replay is
	// being executed, if any.
	{
		err := config.Trial(ctx)
		if err != nil {
			return microerror.MaskAny(err)
		}

		key := transactionKey("transaction", transactionID, "trial", config.TrialID)
		err = e.storage.Create(ctx, key, "{}")
		if err != nil {
			return microerror.MaskAny(err)
		}

		e.logger.Log("debug", fmt.Sprintf("executed transaction trial for transaction ID %s and trial ID %s", transactionID, config.TrialID))
	}

	return nil
}

func (e *executer) ExecuteConfig() ExecuteConfig {
	return ExecuteConfig{
		Replay:  nil,
		Trial:   nil,
		TrialID: "",
	}
}

func validateExecuteConfig(config ExecuteConfig) error {
	if config.Trial == nil {
		return microerror.MaskAnyf(invalidExecutionError, "trial must not be empty")
	}
	if config.TrialID == "" {
		return microerror.MaskAnyf(invalidExecutionError, "trial ID must not be empty")
	}

	return nil
}
