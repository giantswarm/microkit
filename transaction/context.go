package transaction

import (
	"context"
	"fmt"
)

type contextKey int

const (
	trackedKey contextKey = iota
	transactionIDKey
)

func WithTracked(ctx context.Context, tracked bool) context.Context {
	return context.WithValue(ctx, trackedKey, tracked)
}

func GetTracked(ctx context.Context) (tracked bool, ok bool) {
	v := ctx.Value(trackedKey)
	if v == nil {
		return false, false
	}
	b, ok := v.(bool)
	if !ok {
		panic(fmt.Sprintf("tracked must be a bool, got %#v.(%T)", v, v))
	}
	return b, true
}

func WithTransactionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, transactionIDKey, id)
}

func GetTransactionID(ctx context.Context) (string, bool) {
	v := ctx.Value(transactionIDKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		panic(fmt.Sprintf("transactionID must be a string, got %#v.(%T)", v, v))
	}
	return s, true
}
