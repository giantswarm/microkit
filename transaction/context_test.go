package transaction

import (
	"context"
	"reflect"
	"testing"
)

func TestContext_Tracked(t *testing.T) {
	ctx := context.Background()

	_, ok := GetTracked(ctx)
	if !reflect.DeepEqual(false, ok) {
		t.Fatalf("expected ok=%#v, got %#v", false, ok)
	}

	ctx = WithTracked(ctx, true)

	tracked, ok := GetTracked(ctx)
	if !reflect.DeepEqual(true, ok) {
		t.Fatalf("expected ok=%#v, got %#v", true, ok)
	}
	if !reflect.DeepEqual(true, tracked) {
		t.Fatalf("expected tracked=%#v, got %#v", true, tracked)
	}

	ctx = WithTracked(ctx, false)

	tracked, ok = GetTracked(ctx)
	if !reflect.DeepEqual(true, ok) {
		t.Fatalf("expected ok=%#v, got %#v", true, ok)
	}
	if !reflect.DeepEqual(false, tracked) {
		t.Fatalf("expected tracked=%#v, got %#v", false, tracked)
	}
}

func TestContext_TransactionID(t *testing.T) {
	ctx := context.Background()

	wid := "test-transaction-id"

	_, ok := GetTransactionID(ctx)
	if !reflect.DeepEqual(false, ok) {
		t.Fatalf("expected ok=%#v, got %#v", "", ok)
	}

	ctx = WithTransactionID(ctx, wid)

	id, ok := GetTransactionID(ctx)
	if !reflect.DeepEqual(true, ok) {
		t.Fatalf("expected ok=%#v, got %#v", true, id)
	}
	if !reflect.DeepEqual(wid, id) {
		t.Fatalf("expected id=%#v, got %#v", wid, id)
	}
}
