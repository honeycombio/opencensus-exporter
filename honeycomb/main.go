package main

import (
	"context"
	"fmt"
	"time"

	libhoney "github.com/honeycombio/libhoney-go"
	"go.opencensus.io/trace"
)

type Exporter struct{}

type Annotation struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

type Span struct {
	TraceID  string `json:"traceId"`
	Name     string `json:"name"`
	ID       string `json:"id"`
	ParentID string `json:"parentId,omitempty"`
	// ServiceName string        `json:"serviceName,omitempty"`
	// HostIPv4    string        `json:"hostIPv4,omitempty"`
	// Port        int           `json:"port,omitempty"`
	DurationMs  time.Duration `json:"durationMs,omitempty"`
	Timestamp   time.Time     `json:"timestamp,omitempty"`
	Annotations []Annotation  `json:"annotations,omitempty"`
}

func (e *Exporter) ExportSpan(sd *trace.SpanData) {
	libhoney.SendNow(honeycombSpan(sd))
}

func honeycombSpan(s *trace.SpanData) Span {
	sc := s.SpanContext
	hcSpan := Span{
		TraceID:   sc.TraceID.String(),
		ID:        sc.SpanID.String(),
		Timestamp: s.StartTime,
	}
	// Hmmm..... it doesn't like this. We're not getting parent spans
	if s.ParentSpanID != (trace.SpanID{}) {
		hcSpan.ParentID = s.ParentSpanID.String()
	}
	if s, e := s.StartTime, s.EndTime; !s.IsZero() && !e.IsZero() {
		hcSpan.DurationMs = e.Sub(s)
	}
	if len(s.Annotations) != 0 || len(s.MessageEvents) != 0 {
		hcSpan.Annotations = make([]Annotation, 0, len(s.Annotations)+len(s.MessageEvents))
		for _, a := range s.Annotations {
			hcSpan.Annotations = append(hcSpan.Annotations, Annotation{
				Timestamp: a.Time,
				Value:     a.Message,
			})
		}
		for _, m := range s.MessageEvents {
			a := Annotation{
				Timestamp: m.Time,
			}
			switch m.EventType {
			case trace.MessageEventTypeSent:
				a.Value = "SENT"
			case trace.MessageEventTypeRecv:
				a.Value = "RECV"
			default:
				a.Value = "<?>"
			}
			hcSpan.Annotations = append(hcSpan.Annotations, a)
		}
	}
	return hcSpan
}

// type customTraceExporter struct{}

// func (ce *customTraceExporter) ExportSpan(sd *trace.SpanData) {
// 	fmt.Printf("Name: %s\nTraceID: %x\nSpanID: %x\nParentSpanID: %x\nStartTime: %s\nEndTime: %s\nAnnotations: %+v\n\n",
// 		sd.Name, sd.TraceID, sd.SpanID, sd.ParentSpanID, sd.StartTime, sd.EndTime, sd.Annotations)
// }

func main() {
	libhoney.Init(libhoney.Config{
		WriteKey: "5ba769dbf42cbd66c950ffc8701c07d2",
		Dataset:  "opencensus-test-integration",
	})
	defer libhoney.Close()

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	trace.RegisterExporter(new(Exporter))
	// trace.RegisterExporter(new(customTraceExporter))

	for i := 0; i < 5; i++ {
		_, span := trace.StartSpan(context.Background(), fmt.Sprintf("sample-%d", i))
		span.Annotate([]trace.Attribute{trace.Int64Attribute("invocations", 1)}, "Invoked it")
		span.End()
		<-time.After(10 * time.Millisecond)
	}
	<-time.After(500 * time.Millisecond)
}
