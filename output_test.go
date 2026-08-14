package manazashi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestWriteErrorJSON(t *testing.T) {
	var output bytes.Buffer
	err := fmt.Errorf("query failed: %w", newIndexNotFoundError("/tmp/index.sqlite", "run rebuild first, or pass --db"))
	if writeErr := WriteError(&output, []string{"files", "--format", "json", "config"}, err); writeErr != nil {
		t.Fatal(writeErr)
	}
	var result cliErrorJSONResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("error JSON = %q: %v", output.String(), err)
	}
	if result.Error.Code != "index_not_found" || result.Error.Message != err.Error() {
		t.Fatalf("error JSON = %#v", result)
	}
	if result.Error.Details["db"] != "/tmp/index.sqlite" {
		t.Fatalf("error details = %#v", result.Error.Details)
	}
}

func TestWriteErrorJSONFallback(t *testing.T) {
	var output bytes.Buffer
	err := errors.New("unexpected failure")
	if writeErr := WriteError(&output, []string{"stats", "--format=json"}, err); writeErr != nil {
		t.Fatal(writeErr)
	}
	var result cliErrorJSONResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("error JSON = %q: %v", output.String(), err)
	}
	if result.Error.Code != "command_failed" || result.Error.Message != "unexpected failure" || result.Error.Details != nil {
		t.Fatalf("error JSON = %#v", result)
	}
}

func TestWriteErrorText(t *testing.T) {
	var output bytes.Buffer
	if err := WriteError(&output, []string{"stats"}, errors.New("unexpected failure")); err != nil {
		t.Fatal(err)
	}
	if output.String() != "unexpected failure\n" {
		t.Fatalf("text error = %q", output.String())
	}
}

func TestRequestsJSONErrorsUsesLastFormatBeforeSeparator(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "separate", args: []string{"files", "--format", "json", "config"}, want: true},
		{name: "equals", args: []string{"files", "--format=json", "config"}, want: true},
		{name: "last format wins", args: []string{"files", "--format", "text", "--format=json"}, want: true},
		{name: "last text wins", args: []string{"files", "--format=json", "--format", "text"}, want: false},
		{name: "after separator", args: []string{"files", "--", "--format=json"}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := requestsJSONErrors(test.args); got != test.want {
				t.Fatalf("requestsJSONErrors(%s) = %t, want %t", strings.Join(test.args, " "), got, test.want)
			}
		})
	}
}
