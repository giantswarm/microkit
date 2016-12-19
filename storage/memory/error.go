package memory

import (
	"github.com/juju/errgo"
)

var keyNotFoundError = errgo.New("key not found")

func IsKeyNotFound(err error) bool {
	return errgo.Cause(err) == keyNotFoundError
}
