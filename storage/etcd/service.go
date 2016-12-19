// Package etcd provides a service that implements an etcd storage.
package etcd

import (
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"

	microerror "github.com/giantswarm/microkit/error"
)

// Config represents the configuration used to create a etcd service.
type Config struct {
	// Dependencies.
	EtcdClient client.Client
}

// DefaultConfig provides a default configuration to create a new etcd service
// by best effort.
func DefaultConfig() Config {
	etcdConfig := client.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Transport: client.DefaultTransport,
	}
	etcdClient, err := client.New(etcdConfig)
	if err != nil {
		panic(err)
	}

	return Config{
		// Dependencies.
		EtcdClient: etcdClient,
	}
}

// New creates a new configured etcd service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.EtcdClient == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "etcd client must not be empty")
	}

	newService := &Service{
		// Dependencies.
		etcdClient: config.EtcdClient,

		// Internals.
		keyClient: client.NewKeysAPI(config.EtcdClient),
	}

	return newService, nil
}

// Service is the etcd service.
type Service struct {
	// Dependencies.
	etcdClient client.Client

	// Internals.
	keyClient client.KeysAPI
}

func (s *Service) Create(key, value string) error {
	_, err := s.keyClient.Create(context.TODO(), key, value)
	if err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s *Service) Delete(key string) error {
	options := &client.DeleteOptions{
		Recursive: true,
	}
	_, err := s.keyClient.Delete(context.TODO(), key, options)
	if err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s *Service) Exists(key string) (bool, error) {
	options := &client.GetOptions{
		Quorum: true,
	}
	_, err := s.keyClient.Get(context.TODO(), key, options)
	if client.IsKeyNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, microerror.MaskAny(err)
	}

	return true, nil
}

func (s *Service) Search(key string) (string, error) {
	options := &client.GetOptions{
		Quorum: true,
	}
	clientResponse, err := s.keyClient.Get(context.TODO(), key, options)
	if client.IsKeyNotFound(err) {
		return "", microerror.MaskAnyf(keyNotFoundError, key)
	} else if err != nil {
		return "", microerror.MaskAny(err)
	}

	return clientResponse.Node.Value, nil
}
