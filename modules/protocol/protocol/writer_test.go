package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestWriterEvents(t *testing.T) {
	tests := []struct {
		name  string
		call  func(*Writer, any) error
		event string
	}{
		{name: "progress", call: (*Writer).Progress, event: "progress"},
		{name: "data", call: (*Writer).Data, event: "data"},
		{name: "result", call: (*Writer).Result, event: "result"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := test.call(NewWriter(&output), map[string]string{"state": "ready"}); err != nil {
				t.Fatalf("writer returned error: %v", err)
			}
			var event Event
			if err := json.Unmarshal(output.Bytes(), &event); err != nil {
				t.Fatalf("decode event: %v", err)
			}
			if event.ProtocolVersion != Version || event.Event != test.event || event.Data == nil {
				t.Fatalf("event = %#v, want version %d and event %q with data", event, Version, test.event)
			}
		})
	}
}

func TestWriterReturnsEncodeError(t *testing.T) {
	writer := NewWriter(errorWriter{})
	for _, call := range []struct {
		name string
		fn   func() error
	}{
		{name: "progress", fn: func() error { return writer.Progress("data") }},
		{name: "data", fn: func() error { return writer.Data("data") }},
		{name: "result", fn: func() error { return writer.Result("data") }},
	} {
		t.Run(call.name, func(t *testing.T) {
			if err := call.fn(); !errors.Is(err, errWriteFailed) {
				t.Fatalf("writer error = %v, want %v", err, errWriteFailed)
			}
		})
	}
}

var errWriteFailed = errors.New("write failed")

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errWriteFailed }
