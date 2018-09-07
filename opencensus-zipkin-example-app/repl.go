package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/trace"

	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

func main() {
	localEndpointURI := "192.168.1.5:5454"
	reporterURI := "http://localhost:9411/api/v2/spans"
	serviceName := "server"

	// In a REPL:
	//   1. Read input
	//   2. process input
	br := bufio.NewReader(os.Stdin)

	localEndpoint, err := openzipkin.NewEndpoint(serviceName, localEndpointURI)
	if err != nil {
		log.Fatalf("Failed to create Zipkin localEndpoint with URI %q error: %v", localEndpointURI, err)
	}

	reporter := zipkinHTTP.NewReporter(reporterURI)
	ze := zipkin.NewExporter(reporter, localEndpoint)

	// And now finally register it as a Trace Exporter
	trace.RegisterExporter(ze)

	// For demo purposes, set the trace sampling probability to be high
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(1.0)})

	// repl is the read, evaluate, print, loop
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
	ctx, span := trace.StartSpan(context.Background(), "repl")
	defer span.End()

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
	ctx, span := trace.StartSpan(ctx, "readLine")
	defer span.End()

	line, _, err := br.ReadLine()
	if err != nil {
		return nil, err
	}

	return line, err
}

// processLine takes in a line of text and
// transforms it. Currently it just capitalizes it.
func processLine(ctx context.Context, in []byte) (out []byte, err error) {
	_, span := trace.StartSpan(ctx, "processLine")
	defer span.End()

	return bytes.ToUpper(in), nil
}
