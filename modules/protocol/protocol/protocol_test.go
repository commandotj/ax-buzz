package protocol

import "testing"

func TestValidateEvent(t *testing.T) {
	tests := []struct {
		name       string
		version    int
		event      string
		hasData    bool
		resultSeen bool
		wantError  string
	}{
		{name: "progress", version: Version, event: "progress", hasData: true},
		{name: "result", version: Version, event: "result", hasData: true},
		{name: "unsupported version", version: Version + 1, event: "progress", hasData: true, wantError: "unsupported Buzz CLI protocol version 2"},
		{name: "after result", version: Version, event: "progress", hasData: true, resultSeen: true, wantError: "Buzz CLI emitted data after its result event"},
		{name: "unknown event", version: Version, event: "error", hasData: true, wantError: `unknown Buzz CLI event "error"`},
		{name: "missing data", version: Version, event: "result", wantError: `Buzz CLI event "result" has no data`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateEvent(test.version, test.event, test.hasData, test.resultSeen)
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("ValidateEvent() error = %q, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("ValidateEvent() error = %v, want %q", err, test.wantError)
			}
		})
	}
}
