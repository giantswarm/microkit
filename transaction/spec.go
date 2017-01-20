package transaction

import (
	"context"
)

// ExecuteConfig is used to configure the Executer.
type ExecuteConfig struct {
	// Replay is the action being executed to replay a transaction result in case
	// a former call to Trial was successful.
	Replay func(context context.Context) error
	// Trial is the action being executed to fulfil a transaction.
	Trial func(context context.Context) error
	// TrialID is an identifier scoped to the transaction ID obtained by the
	// context provided to transaction.Executer.Execute. The trial ID is used to
	// keep track of the state of the current transaction.
	TrialID string
}

// Executer provides a single transactional execution according to the provided
// configuration.
type Executer interface {
	// Execute actually executes the configured transaction. Transactions are
	// identified by the transaction ID obtained by the given context. In case the
	// given context contains a transaction ID, the configured trial associated
	// with the given trial ID will only be executed successfully once for the
	// current transaction ID. In case the trial fails it will be executed again
	// on the next call to Execute. In case the trial succeeded the given deplay
	// function, if any, will be executed.
	Execute(ctx context.Context, config ExecuteConfig) error
	// ExecuteConfig provides a default configuration for calls to Execute by best
	// effort.
	ExecuteConfig() ExecuteConfig
}
