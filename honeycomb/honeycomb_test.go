package honeycomb

import (
	"context"
	"reflect"
	"testing"
	"time"

	libhoney "github.com/honeycombio/libhoney-go"
	"github.com/stretchr/testify/assert"
	"go.opencensus.io/trace"
)

func TestExport(t *testing.T) {
	now := time.Now().Round(time.Microsecond)
	tests := []struct {
		span *trace.SpanData
		want Span
	}{
		{
			span: &trace.SpanData{
				SpanContext: trace.SpanContext{
					TraceID:      trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					SpanID:       trace.SpanID{17, 18, 19, 20, 21, 22, 23, 24},
					TraceOptions: 1,
				},
				Name:      "name",
				SpanKind:  trace.SpanKindClient,
				StartTime: now,
				EndTime:   now.Add(time.Duration(0.5 * float64(time.Millisecond))),
				Attributes: map[string]interface{}{
					"stringkey": "value",
					"intkey":    int64(42),
					"boolkey1":  true,
					"boolkey2":  false,
				},
				MessageEvents: []trace.MessageEvent{
					{
						Time:                 now,
						EventType:            trace.MessageEventTypeSent,
						MessageID:            12,
						UncompressedByteSize: 99,
						CompressedByteSize:   98,
					},
				},
				Annotations: []trace.Annotation{
					{
						Time:    now,
						Message: "Annotation",
						Attributes: map[string]interface{}{
							"stringkey": "value",
							"intkey":    int64(42),
							"boolkey1":  true,
							"boolkey2":  false,
						},
					},
				},
				Status: trace.Status{
					Code:    3,
					Message: "error",
				},
			},
			want: Span{
				TraceID:    "0102030405060708090a0b0c0d0e0f10",
				ID:         "1112131415161718",
				Name:       "name",
				ParentID:   "",
				Timestamp:  now,
				DurationMs: 0.5,
				Annotations: []Annotation{
					{
						Timestamp: now,
						Value:     "Annotation",
					},
					{
						Timestamp: now,
						Value:     "SENT",
					},
				},
			},
		},
		{
			span: &trace.SpanData{
				SpanContext: trace.SpanContext{
					TraceID:      trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					SpanID:       trace.SpanID{17, 18, 19, 20, 21, 22, 23, 24},
					TraceOptions: 1,
				},
				Name:      "name",
				StartTime: now,
				EndTime:   now.Add(24 * time.Hour),
			},
			want: Span{
				TraceID:    "0102030405060708090a0b0c0d0e0f10",
				ID:         "1112131415161718",
				Name:       "name",
				Timestamp:  now,
				DurationMs: 86400000,
			},
		},
		{
			span: &trace.SpanData{
				SpanContext: trace.SpanContext{
					TraceID:      trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					SpanID:       trace.SpanID{17, 18, 19, 20, 21, 22, 23, 24},
					TraceOptions: 1,
				},
				Name:      "name",
				StartTime: now,
				EndTime:   now.Add(24 * time.Hour),
				Status: trace.Status{
					Code:    0,
					Message: "there is no cause for alarm",
				},
			},
			want: Span{

				TraceID:    "0102030405060708090a0b0c0d0e0f10",
				ID:         "1112131415161718",
				Name:       "name",
				Timestamp:  now,
				DurationMs: 86400000,
			},
		},
		{
			span: &trace.SpanData{
				SpanContext: trace.SpanContext{
					TraceID:      trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					SpanID:       trace.SpanID{17, 18, 19, 20, 21, 22, 23, 24},
					TraceOptions: 1,
				},
				Name:      "name",
				StartTime: now,
				EndTime:   now.Add(24 * time.Hour),
				Status: trace.Status{
					Code: 1234,
				},
			},
			want: Span{
				TraceID:    "0102030405060708090a0b0c0d0e0f10",
				ID:         "1112131415161718",
				Name:       "name",
				Timestamp:  now,
				DurationMs: 86400000,
			},
		},
	}
	for _, tt := range tests {
		got := honeycombSpan(tt.span)
		if len(got.Annotations) != len(tt.want.Annotations) {
			t.Fatalf("honeycombSpan: got %d annotations in span, want %d", len(got.Annotations), len(tt.want.Annotations))
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("honeycombSpan:\n\tgot  %#v\n\twant %#v", got, tt.want)
		}
	}
}

func TestHoneycombOutput(t *testing.T) {
	mockHoneycomb := &libhoney.MockOutput{}
	assert := assert.New(t)
	exporter := NewExporter("overridden", "overridden")
	exporter.ServiceName = "honeycomb-test"

	libhoney.Init(libhoney.Config{
		WriteKey: "test",
		Dataset:  "test",
		Output:   mockHoneycomb,
	})
	exporter.Builder = libhoney.NewBuilder()

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	_, span := trace.StartSpan(context.TODO(), "mySpan")
	time.Sleep(time.Duration(0.5 * float64(time.Millisecond)))
	span.AddAttributes(trace.StringAttribute("attributeName", "attributeValue"))
	span.End()

	assert.Equal(1, len(mockHoneycomb.Events()))
	traceID := mockHoneycomb.Events()[0].Fields()["trace.trace_id"]
	assert.Equal(span.SpanContext().TraceID.String(), traceID)

	spanID := mockHoneycomb.Events()[0].Fields()["trace.span_id"]
	assert.Equal(span.SpanContext().SpanID.String(), spanID)

	name := mockHoneycomb.Events()[0].Fields()["name"]
	assert.Equal("mySpan", name)

	durationMs := mockHoneycomb.Events()[0].Fields()["duration_ms"]
	durationMsFl, ok := durationMs.(float64)
	assert.Equal(ok, true)
	assert.Equal((durationMsFl > 0), true)
	assert.Equal((durationMsFl < 1), true)

	attributeName := mockHoneycomb.Events()[0].Fields()["attributeName"]
	assert.Equal("attributeValue", attributeName)

	timestamp := mockHoneycomb.Events()[0].Fields()["timestamp"]
	assert.Equal(mockHoneycomb.Events()[0].Timestamp, timestamp)

	serviceName := mockHoneycomb.Events()[0].Fields()["service_name"]
	assert.Equal("honeycomb-test", serviceName)
	assert.Equal(mockHoneycomb.Events()[0].Dataset, "test")
}
