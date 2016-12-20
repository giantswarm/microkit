package memory

import (
	"testing"
)

func Test_Service(t *testing.T) {
	newStorage, err := New(DefaultConfig())
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	key := "test-key"
	value := "test-value"

	ok, err := newStorage.Exists(key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if ok {
		t.Fatal("expected", false, "got", true)
	}

	err = newStorage.Create(key, value)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	ok, err = newStorage.Exists(key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if !ok {
		t.Fatal("expected", true, "got", false)
	}

	v, err := newStorage.Search(key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if v != value {
		t.Fatal("expected", value, "got", v)
	}

	err = newStorage.Delete(key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	ok, err = newStorage.Exists(key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if ok {
		t.Fatal("expected", false, "got", true)
	}

	v, err = newStorage.Search(key)
	if !IsKeyNotFound(err) {
		t.Fatal("expected", true, "got", false)
	}
	if v != "" {
		t.Fatal("expected", "empty string", "got", v)
	}
}
