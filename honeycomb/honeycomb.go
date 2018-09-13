package honeycomb

import (
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
		Name:      s.Name,
		Timestamp: s.StartTime,
	}

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
