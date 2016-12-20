package etcd

import (
	"github.com/juju/errgo"
)

var invalidConfigError = errgo.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}

var keyNotFoundError = errgo.New("key not found")

// IsKeyNotFound asserts keyNotFoundError.
func IsKeyNotFound(err error) bool {
	return errgo.Cause(err) == keyNotFoundError
}
