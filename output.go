package manazashi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type outputFormat string

const (
	outputFormatText outputFormat = "text"
	outputFormatJSON outputFormat = "json"
)

func parseOutputFormat(value string) (outputFormat, error) {
	format := outputFormat(value)
	switch format {
	case outputFormatText, outputFormatJSON:
		return format, nil
	default:
		return "", fmt.Errorf("unsupported output format %q: use text or json", value)
	}
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

type cliError struct {
	code    string
	message string
	details map[string]any
}

func (e *cliError) Error() string {
	return e.message
}

type cliErrorJSON struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type cliErrorJSONResult struct {
	Error cliErrorJSON `json:"error"`
}

func newIndexNotFoundError(db, guidance string) error {
	return &cliError{
		code:    "index_not_found",
		message: fmt.Sprintf("index not found: %s; %s", db, guidance),
		details: map[string]any{"db": db},
	}
}

// WriteError writes err using the output format explicitly requested by args.
func WriteError(w io.Writer, args []string, err error) error {
	if !requestsJSONErrors(args) {
		_, writeErr := fmt.Fprintln(w, err)
		return writeErr
	}
	payload := cliErrorJSON{Code: "command_failed", Message: err.Error()}
	var commandErr *cliError
	if errors.As(err, &commandErr) {
		payload.Code = commandErr.code
		payload.Details = commandErr.details
	}
	return writeJSON(w, cliErrorJSONResult{Error: payload})
}

func requestsJSONErrors(args []string) bool {
	format := ""
	for index := 1; index < len(args); index++ {
		arg := args[index]
		if arg == "--" {
			break
		}
		if arg == "--format" && index+1 < len(args) {
			format = args[index+1]
			index++
			continue
		}
		if strings.HasPrefix(arg, "--format=") {
			format = strings.TrimPrefix(arg, "--format=")
		}
	}
	return format == string(outputFormatJSON)
}
