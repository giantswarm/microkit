// +build integration

package etcd

import (
	"testing"
)

func Test_CreateExistsSearch(t *testing.T) {
	config := DefaultConfig()
	config.Prefix = "foo"
	newStorage, err := New(config)
	if err != nil {
		panic(err)
	}

	key := "my-key"
	val := "my-val"

	// There should be no key/value pair being stored initially.
	ok, err := newStorage.Exists(key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if ok {
		t.Fatal("expected", false, "got", true)
	}

	// Creating the key/value pair should work.
	err = newStorage.Create(key, val)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	// There should be the created key/value pair.
	ok, err = newStorage.Exists(key)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if !ok {
		t.Fatal("expected", true, "got", false)
	}
}

func Test_List(t *testing.T) {
	config := DefaultConfig()
	config.Prefix = "foo"
	newStorage, err := New(config)
	if err != nil {
		panic(err)
	}

	val := "my-val"

	err = newStorage.Create("key/one", val)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	err = newStorage.Create("key/two", val)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	values, err := newStorage.List("key")
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	if len(values) != 2 {
		t.Fatal("expected", 2, "got", len(values))
	}
}

func Test_List_Invalid(t *testing.T) {
	config := DefaultConfig()
	config.Prefix = "foo"
	newStorage, err := New(config)
	if err != nil {
		panic(err)
	}

	val := "my-val"

	err = newStorage.Create("key/one", val)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}
	err = newStorage.Create("key/two", val)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	_, err = newStorage.List("ke")
	if !IsNotFound(err) {
		t.Fatal("expected", true, "got", false)
	}
}
