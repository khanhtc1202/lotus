/*

Generated by using code-generator

*/

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/nghialv/lotus/pkg/app/lotus/apis/lotus/v1beta1"
	"github.com/nghialv/lotus/pkg/app/lotus/client/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type LotusV1beta1Interface interface {
	RESTClient() rest.Interface
	LotusesGetter
}

// LotusV1beta1Client is used to interact with features provided by the lotus.nghialv.com group.
type LotusV1beta1Client struct {
	restClient rest.Interface
}

func (c *LotusV1beta1Client) Lotuses(namespace string) LotusInterface {
	return newLotuses(c, namespace)
}

// NewForConfig creates a new LotusV1beta1Client for the given config.
func NewForConfig(c *rest.Config) (*LotusV1beta1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &LotusV1beta1Client{client}, nil
}

// NewForConfigOrDie creates a new LotusV1beta1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *LotusV1beta1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new LotusV1beta1Client for the given RESTClient.
func New(c rest.Interface) *LotusV1beta1Client {
	return &LotusV1beta1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1beta1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *LotusV1beta1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
