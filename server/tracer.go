package server

import (
	"io"

	"github.com/giantswarm/microerror"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func tracer(name string) (opentracing.Tracer, io.Closer, error) {
	c := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}

	// TODO: Integrate with micrologger.
	tracer, closer, err := c.New(name, config.Logger(jaeger.StdLogger))
	if err != nil {
		return nil, nil, microerror.Mask(err)
	}

	return tracer, closer, nil
}
