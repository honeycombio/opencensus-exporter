package honeycomb

import (
	"time"

	libhoney "github.com/honeycombio/libhoney-go"
	"go.opencensus.io/trace"
)

type Exporter struct {
	Builder *libhoney.Builder
}

type Annotation struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

type Span struct {
	TraceID     string       `json:"traceId"`
	Name        string       `json:"name"`
	ID          string       `json:"id"`
	ParentID    string       `json:"parentId,omitempty"`
	DurationMs  int          `json:"durationMs,omitempty"`
	Timestamp   time.Time    `json:"timestamp,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
}

func (e *Exporter) Close() {
	libhoney.Close()
}

func NewExporter(writeKey, dataset string) *Exporter {
	libhoney.Init(libhoney.Config{
		WriteKey: writeKey,
		Dataset:  dataset,
	})
	return &Exporter{Builder: libhoney.NewBuilder()}
}

func (e *Exporter) ExportSpan(sd *trace.SpanData) {
	ev := e.Builder.NewEvent()
	ev.Timestamp = sd.StartTime
	hs := honeycombSpan(sd)
	ev.Add(hs)

	// Add an event field for each attribute
	if len(sd.Attributes) != 0 {
		for key, value := range sd.Attributes {
			switch v := value.(type) {
			case bool:
				if v {
					ev.AddField(key, true)
				} else {
					ev.AddField(key, false)
				}
			default:
				ev.AddField(key, v)
			}
		}
	}

	// Add an event field for status code and status message
	if sd.Status.Code != 0 || sd.Status.Message != "" {
		if sd.Status.Code != 0 {
			ev.AddField("status_code", sd.Status.Code)
		}
		if sd.Status.Message != "" {
			ev.AddField("status_description", sd.Status.Message)
		}
	}
	ev.Send()
}

func honeycombSpan(s *trace.SpanData) Span {
	sc := s.SpanContext
	hcSpan := Span{
		TraceID:   sc.TraceID.String(),
		ID:        sc.SpanID.String(),
		Name:      s.Name,
		Timestamp: s.StartTime,
	}

	if s.ParentSpanID != (trace.SpanID{}) {
		hcSpan.ParentID = s.ParentSpanID.String()
	}

	if s, e := s.StartTime, s.EndTime; !s.IsZero() && !e.IsZero() {
		hcSpan.DurationMs = int(e.Sub(s) / time.Millisecond)
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
