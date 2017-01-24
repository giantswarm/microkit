package server

import (
	"strings"
)

func transactionKey(keys ...string) string {
	return strings.Join(keys, "/")
}
