package honeycomb

import (
	"reflect"
	"testing"
	"time"

	"go.opencensus.io/trace"
)

func TestExport(t *testing.T) {
	// Since Zipkin reports in microsecond resolution let's round our Timestamp,
	// so when deserializing Zipkin data in this test we can properly compare.
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
				DurationMs: (now.Add(24 * time.Hour).Sub(now)),
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

		// {TraceID:"0102030405060708090a0b0c0d0e0f10",
		// Name:"name",
		// ID:"1112131415161718",
		// ParentID:"",
		// DurationMs:86400000000000, 
		// Timestamp:time.Time{wall:0x23999238, ext:63672544799, loc:(*time.Location)(0x1497f80)}, 
		// Annotations:[]honeycomb.Annotation{honeycomb.Annotation{Timestamp:time.Time{wall:0x23999238, ext:63672544799, loc:(*time.Location)(0x1497f80)}, Value:"Annotation"}, honeycomb.Annotation{Timestamp:time.Time{wall:0x23999238, ext:63672544799, loc:(*time.Location)(0x1497f80)}, Value:"SENT"}}}

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
				ID:"1112131415161718",
				Name:       "name",
				Timestamp:  now,
				DurationMs: (now.Add(24 * time.Hour).Sub(now)),
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
				ID:"1112131415161718",
				Name:       "name",
				Timestamp:  now,
				DurationMs: (now.Add(24 * time.Hour).Sub(now)),
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
				ID:"1112131415161718",
				Name:       "name",
				Timestamp:  now,
				DurationMs: (now.Add(24 * time.Hour).Sub(now)),
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
