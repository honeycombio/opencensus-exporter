package honeycomb

import (
	"reflect"
	"testing"
	"time"

	"go.opencensus.io/trace"
)

func getTimeinMs(dur time.Duration) int {
	return int(dur.Nanoseconds()) / int(time.Millisecond)
}

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
				EndTime:   now.Add(24 * time.Hour),
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
				DurationMs: getTimeinMs(now.Add(24 * time.Hour).Sub(now)),
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
				DurationMs: getTimeinMs(now.Add(24 * time.Hour).Sub(now)),
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
				DurationMs: getTimeinMs(now.Add(24 * time.Hour).Sub(now)),
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
				DurationMs: getTimeinMs(now.Add(24 * time.Hour).Sub(now)),
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
