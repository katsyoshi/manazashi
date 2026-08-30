package symbols

import (
	"strings"
	"testing"
)

func TestExtractGoSymbolsUsesAST(t *testing.T) {
	source := `package sample

const answer = 42

var (
	count int
	ignored = func() {}
)

type Service struct{}
type Alias = string

func NewService() *Service {
	return &Service{}
}

func (s *Service) Run(ctx context.Context) error {
	return nil
}
`
	symbols := Extract("sample.go", "go", splitLines(source))

	assertSymbol(t, symbols, "constant", "answer", 3)
	assertSymbol(t, symbols, "variable", "count", 6)
	assertSymbol(t, symbols, "variable", "ignored", 7)
	assertSymbol(t, symbols, "type", "Service", 10)
	assertSymbol(t, symbols, "type", "Alias", 11)
	assertSymbol(t, symbols, "function", "NewService", 13)
	assertSymbol(t, symbols, "method", "Run", 17)
	assertSymbolEndLine(t, symbols, "function", "NewService", 15)
	assertSymbolEndLine(t, symbols, "method", "Run", 19)
	assertNoSymbol(t, symbols, "function", "ignored")
}

func TestExtractGoSymbolsHandlesPartialParse(t *testing.T) {
	source := `package sample

func stillIndexed() {
`
	symbols := Extract("broken.go", "go", splitLines(source))

	assertSymbol(t, symbols, "function", "stillIndexed", 3)
}

func TestExtractGoSymbolsUsesPhysicalPositions(t *testing.T) {
	source := `package sample

func before() {}

//line generated.go:999999
func after() {
}
`
	symbols := Extract("sample.go", "go", splitLines(source))

	assertSymbol(t, symbols, "function", "before", 3)
	assertSymbol(t, symbols, "function", "after", 6)
	assertSymbolEndLine(t, symbols, "function", "after", 7)
	for _, symbol := range symbols {
		if symbol.Kind == "function" && symbol.Name == "after" {
			if symbol.Signature != "func after() {" {
				t.Fatalf("symbol signature = %q, want %q", symbol.Signature, "func after() {")
			}
			if !strings.Contains(symbol.Context, "func after()") {
				t.Fatalf("symbol context = %q, want physical source context", symbol.Context)
			}
			return
		}
	}
	t.Fatal("function after not found")
}

func TestSymbolContextHandlesOutOfRangeLine(t *testing.T) {
	lines := []string{"first", "second"}
	for _, line := range []int{-1, 0, 3, 999999} {
		if context := symbolContext(lines, line); context != "" {
			t.Fatalf("symbolContext(%d) = %q, want empty context", line, context)
		}
	}
}

func TestExtractMacroSymbols(t *testing.T) {
	tests := []struct {
		language string
		source   string
		name     string
	}{
		{language: "c", source: "#define DECLARE_CLASS(name) class name {}", name: "DECLARE_CLASS"},
		{language: "cpp", source: "# define VALUE 42", name: "VALUE"},
		{language: "rust", source: "macro_rules! make_item { () => {} }", name: "make_item"},
		{language: "elisp", source: "(defmacro with-value (name) nil)", name: "with-value"},
		{language: "clojure", source: "(defmacro with-value [name] nil)", name: "with-value"},
		{language: "elixir", source: "defmacro build(value) do", name: "build"},
	}
	for _, test := range tests {
		t.Run(test.language, func(t *testing.T) {
			symbols := Extract("sample", test.language, []string{test.source})
			assertSymbol(t, symbols, "macro", test.name, 1)
		})
	}
}

func assertSymbol(t *testing.T, symbols []Symbol, kind, name string, line int) {
	t.Helper()
	for _, symbol := range symbols {
		if symbol.Kind == kind && symbol.Name == name && symbol.Line == line {
			return
		}
	}
	t.Fatalf("symbol %s %s at line %d not found in %+v", kind, name, line, symbols)
}

func assertNoSymbol(t *testing.T, symbols []Symbol, kind, name string) {
	t.Helper()
	for _, symbol := range symbols {
		if symbol.Kind == kind && symbol.Name == name {
			t.Fatalf("unexpected symbol %s %s found in %+v", kind, name, symbols)
		}
	}
}

func assertSymbolEndLine(t *testing.T, symbols []Symbol, kind, name string, endLine int) {
	t.Helper()
	for _, symbol := range symbols {
		if symbol.Kind == kind && symbol.Name == name {
			if symbol.EndLine != endLine {
				t.Fatalf("symbol %s %s end line = %d, want %d", kind, name, symbol.EndLine, endLine)
			}
			return
		}
	}
	t.Fatalf("symbol %s %s not found in %+v", kind, name, symbols)
}

func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
