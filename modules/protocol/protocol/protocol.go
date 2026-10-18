package protocol

import (
	"encoding/json"
	"fmt"
	"io"
)

const Version = 1

type Event struct {
	ProtocolVersion int    `json:"protocolVersion"`
	Event           string `json:"event"`
	Data            any    `json:"data"`
}

type Writer struct {
	encoder *json.Encoder
}

func NewWriter(output io.Writer) *Writer {
	return &Writer{encoder: json.NewEncoder(output)}
}

func (w *Writer) Progress(data any) error {
	return w.encoder.Encode(Event{ProtocolVersion: Version, Event: "progress", Data: data})
}

func (w *Writer) Data(data any) error {
	return w.encoder.Encode(Event{ProtocolVersion: Version, Event: "data", Data: data})
}

func (w *Writer) Result(data any) error {
	return w.encoder.Encode(Event{ProtocolVersion: Version, Event: "result", Data: data})
}

func ValidateEvent(version int, event string, hasData bool, resultSeen bool) error {
	if version != Version {
		return fmt.Errorf("unsupported Buzz CLI protocol version %d", version)
	}
	if resultSeen {
		return fmt.Errorf("Buzz CLI emitted data after its result event")
	}
	if event != "progress" && event != "data" && event != "result" {
		return fmt.Errorf("unknown Buzz CLI event %q", event)
	}
	if !hasData {
		return fmt.Errorf("Buzz CLI event %q has no data", event)
	}
	return nil
}
