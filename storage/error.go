package storage

import (
	"github.com/juju/errgo"

	"github.com/giantswarm/microkit/storage/etcd"
	"github.com/giantswarm/microkit/storage/memory"
)

var invalidConfigError = errgo.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}

// IsKeyNotFound represents the error matcher for public use. Services using the
// storage service internally should use this public key matcher to verify if
// some storage error is of type "key not found", instead of using a specific
// error matching of some specific storage implementation. This public error
// matcher groups all necessary error matchers of more specific storage
// implementations.
func IsKeyNotFound(err error) bool {
	return etcd.IsKeyNotFound(err) || memory.IsKeyNotFound(err)
}
