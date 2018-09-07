package main

import (
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport/zipkin"
)

func Xmain() {
	backendURI := "http://localhost:9411/api/v1/spans"
	// zipkin.NewHTTPTransport always returns a nil error, no point in checking
	// it. Spans will just be dropped if the backend turns out to be
	// unavailable, which is the behavior we want.
	transport, _ := zipkin.NewHTTPTransport(backendURI)
	reporter := jaeger.NewRemoteReporter(transport)
	sampler := jaeger.NewConstSampler(true)

	zstracer, zscloser := jaeger.NewTracer("opentracing-events", sampler, reporter)

	opentracing.SetGlobalTracer(zstracer)
	defer zscloser.Close()

	helloWorld()
}

func helloWorld() {
	parent := opentracing.GlobalTracer().StartSpan("hello")
	defer parent.Finish()
	child := opentracing.GlobalTracer().StartSpan(
		"world", opentracing.ChildOf(parent.Context()))
	defer child.Finish()
}
