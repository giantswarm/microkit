package server

import (
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/juju/errgo"
)

var invalidConfigError = errgo.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}

func errorDomain(err error) string {
	switch e := err.(type) {
	case kithttp.Error:
		switch e.Domain {
		case kithttp.DomainEncode:
			return "encode"
		case kithttp.DomainDecode:
			return "decode"
		case kithttp.DomainDo:
			return "domain"
		}
	}
	return "server"
}

func errorTrace(err error) string {
	switch kitErr := err.(type) {
	case kithttp.Error:
		switch errgoErr := kitErr.Err.(type) {
		case *errgo.Err:
			return errgoErr.GoString()
		}
	}
	return "n/a"
}

func errorMessage(err error) string {
	switch kitErr := err.(type) {
	case kithttp.Error:
		switch errgoErr := kitErr.Err.(type) {
		case *errgo.Err:
			return errgoErr.Error()
		}
	}
	return err.Error()
}
