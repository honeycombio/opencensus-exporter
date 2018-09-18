package honeycomb

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
				DurationMs: 86400000,
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
	libhoney.Init(libhoney.Config{
		WriteKey: "test",
		Dataset:  "test",
		Output:   mockHoneycomb,
	})
	exporter := Exporter{libhoney.NewBuilder()}

	jsonPayload := `[{
				"traceId":     "350565b6a90d4c8c",
				"name":        "persist",
				"id":          "34472e70cb669b31",
				"parentId":    "",
				"timestamp":  1506629747288651,
				"duration": 192
			}]`

	w := handleGzipped(exporter, []byte(jsonPayload))

	assert.Equal(http.StatusAccepted, w.Code)
	assert.Equal(1, len(mockHoneycomb.Events()))
	assert.Equal(map[string]interface{}{
		"traceId":        "350565b6a90d4c8c",
		"name":           "persist",
		"id":             "34472e70cb669b31",
		"serviceName":    "poodle",
		"hostIPv4":       "10.129.211.111",
		"lc":             "poodle",
		"responseLength": int64(136),
		"durationMs":     0.192,
	}, mockHoneycomb.Events()[0].Fields())
	assert.Equal(mockHoneycomb.Events()[0].Dataset, "test")
}

func handleGzipped(e Exporter, payload []byte) *httptest.ResponseRecorder {
	var compressedPayload bytes.Buffer
	zw := gzip.NewWriter(&compressedPayload)
	zw.Write(payload)
	zw.Close()

	r := httptest.NewRequest("POST", "/api/v1/spans",
		&compressedPayload)
	r.Header.Add("Content-Encoding", "gzip")
	r.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	w = unGzip(w, r)
	return w
}

func unGzip(w *httptest.ResponseRecorder, r *http.Request) *httptest.ResponseRecorder {
	var newBody io.ReadCloser
	isGzipped := r.Header.Get("Content-Encoding")
	if isGzipped == "gzip" {
		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, r.Body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error allocating buffer for ungzipping"))
			return w
		}
		var err error
		newBody, err = gzip.NewReader(&buf)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error ungzipping span data"))
			return w
		}
		r.Body = newBody
	}
	return w
}
