package transaction

import (
	"bytes"
	"net/http"

	"golang.org/x/net/context"
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

// Responder is able to reply to requests for which transactions have been
// tracked.
type Responder interface {
	// Track
	Reply(ctx context.Context, rr ResponseReplier) error
	Track(ctx context.Context, rt ResponseTracker) error
}

// ResponseReplier TODO
type ResponseReplier interface {
	// Header is only a wrapper around http.ResponseWriter.Header.
	Header() http.Header
	// Write is only a wrapper around http.ResponseWriter.Write.
	Write(b []byte) (int, error)
	// WriteHeader is a wrapper around http.ResponseWriter.Write. In addition to
	// that it is used to track the written status code.
	WriteHeader(c int)
}

// ResponseTracker TODO
type ResponseTracker interface {
	// BodyBuffer returns the buffer which is used to track the bytes being
	// written to the response.
	BodyBuffer() *bytes.Buffer
	// Header is only a wrapper around http.ResponseWriter.Header.
	Header() http.Header
	// StatusCode returns either the default status code of the one that was
	// actually written using WriteHeader.
	StatusCode() int
}
