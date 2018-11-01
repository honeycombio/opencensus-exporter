package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/honeycombio/opencensus-exporter/honeycomb"
	"go.opencensus.io/trace"
)

func main() {
	exporter := honeycomb.NewExporter("YOUR-HONEYCOMB-WRITE-KEY", "YOUR-DATASET-NAME")
	defer exporter.Close()

	trace.RegisterExporter(exporter)

	br := bufio.NewReader(os.Stdin)

	sampleFraction := 0.5
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(sampleFraction)})
	// If you use the Open Census Probability Sampler, be sure to pass that sampleFraction to the exporter
	// so that Honeycomb can pick it up and make sure we handle your sampling properly.
	// Note: The Probability Sampler uses a fraction, whereas Honeycomb uses an integer, which is the inverse of that fraction.
	exporter.SampleFraction = sampleFraction

	exporter.ServiceName = "honeycomb-example"

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
