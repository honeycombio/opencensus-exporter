package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport/zipkin"
)

func main() {
	// In a REPL:
	//   1. Read input
	//   2. process input
	br := bufio.NewReader(os.Stdin)

	backendURI := "http://localhost:9411/api/v1/spans"
	// zipkin.NewHTTPTransport always returns a nil error, no point in checking
	// it. Spans will just be dropped if the backend turns out to be
	// unavailable, which is the behavior we want.
	transport, _ := zipkin.NewHTTPTransport(backendURI)
	reporter := jaeger.NewRemoteReporter(transport)
	sampler := jaeger.NewConstSampler(true)

	zstracer, zscloser := jaeger.NewTracer("opentracing-events", sampler, reporter)
	opentracing.SetGlobalTracer(zstracer)

	opentracing.SetGlobalTracer(zstracer)
	defer zscloser.Close()

	// // repl is the read, evaluate, print, loop
	for {
		if err := readEvaluateProcess(br); err != nil {
			if err == io.EOF {
				return
			}
			log.Fatal(err)
		}
	}
}

// readEvaluateProcess reads a line from the input reader and
// then processes it. It returns an error if any was encountered.
func readEvaluateProcess(br *bufio.Reader) error {
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "repl")
	defer span.Finish()

	fmt.Printf("> ")

	line, err := readLine(ctx, br)
	if err != nil {
		return err
	}

	out, err := processLine(ctx, line)
	if err != nil {
		return err
	}
	fmt.Printf("< %s\n\n", out)
	return nil
}

func readLine(ctx context.Context, br *bufio.Reader) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "readLine")
	defer span.Finish()

	line, _, err := br.ReadLine()
	if err != nil {
		return nil, err
	}

	return line, err
}

// processLine takes in a line of text and
// transforms it. Currently it just capitalizes it.
func processLine(ctx context.Context, in []byte) (out []byte, err error) {
	span, _ := opentracing.StartSpanFromContext(context.Background(), "processLine")
	defer span.Finish()

	return bytes.ToUpper(in), nil
}
