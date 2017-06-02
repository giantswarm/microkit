// Package kubernetes provides a service that implements a Kubernetes storage.
package kubernetes

import (
	"fmt"

	microerror "github.com/giantswarm/microkit/error"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

const (
	Endpoint = "apis/giantswarm.io/v1/namespaces/%s/clusters"
)

// Config represents the configuration used to create a Kubernetes service.
type Config struct {
	// Dependencies.
	KubernetesClient kubernetes.Interface

	// Settings.
	Namespace string
}

// DefaultConfig provides a default configuration to create a new Kubernetes
// service by best effort.
func DefaultConfig() Config {
	kubernetesConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	kubernetesClient, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		panic(err)
	}

	return Config{
		// Dependencies.
		KubernetesClient: kubernetesClient,

		// Settings.
		Namespace: "default",
	}
}

// New creates a new configured Kubernetes service.
func New(config Config) (*Service, error) {
	// Dependencies
	if config.KubernetesClient == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "kubernetes client must not be empty")
	}

	newService := &Service{
		// Dependencies.
		kubernetesClient: config.KubernetesClient,

		// Settings.
		namespace: config.Namespace,
	}

	if err := newService.createTPR(); err != nil {
		return nil, microerror.MaskAny(err)
	}

	return newService, nil
}

// Service is the Kubernetes service.
type Service struct {
	// Dependencies
	kubernetesClient kubernetes.Interface

	// Settings
	namespace string
}

func (s *Service) Create(ctx context.Context, key, value string) error {
	restClient := s.kubernetesClient.Core().RESTClient()
	response := restClient.Post().Body(customObject).AbsPath(fmt.Sprintf(Endpoint, s.namespace)).Do()

	err := response.Error()
	if err != nil {
		return microerror.MaskAny(err)
	}

	return nil
}

func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (s *Service) List(ctx context.Context, key string) ([]string, error) {
	return []string{}, nil
}

func (s *Service) Search(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (s *Service) createTPR() error {
	tpr := &v1beta1.ThirdPartyResource{
		ObjectMeta: v1.ObjectMeta{
			Name: "microkit-storage",
		},
		Versions: []v1beta1.APIVersion{
			{Name: "v1"},
		},
		Description: "Generic storage for microkit",
	}

	_, err := s.kubernetesClient.Extensions().ThirdPartyResources().Create(tpr)
	if err != nil && !errors.IsAlreadyExists(err) {
		return microerror.MaskAny(err)
	}

	return nil
}
