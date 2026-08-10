package symbols

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestRegexRubySymbolsFallback(t *testing.T) {
	source := `module Sample
  class Worker
    def perform!
    end
  end
end
`
	symbols, ok := regexSymbolExtractor{}.extract("worker.rb", "ruby", splitLines(source))
	if !ok {
		t.Fatal("regex ruby extraction failed")
	}

	assertSymbol(t, symbols, "module", "Sample", 1)
	assertSymbol(t, symbols, "class", "Worker", 2)
	assertSymbol(t, symbols, "method", "perform", 3)
}

func TestRubyBatchMatchesSingleFileExtraction(t *testing.T) {
	source := `module Sample
  class Worker
    CONST = 1

    def self.build
    end

    def quoter.quote(value)
    end

    def face_with_to_a.to_a
    end

    def perform!
    end
  end
end
`
	path := filepath.Join(t.TempDir(), "worker.rb")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	batch, ok := ExtractRubyBatch([]RubyBatchFile{{Path: "worker.rb", SourcePath: path}})
	if !ok {
		t.Skip("RubyVM::AbstractSyntaxTree batch extraction is unavailable")
	}
	lines := splitLines(source)
	got := ExtractWithRubyBatch(batch, "worker.rb", "ruby", lines)
	want := Extract("worker.rb", "ruby", lines)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("batch symbols = %#v, want %#v", got, want)
	}
}

func TestRubyBatchIgnoresQualifiedConstantWritesLikeSingleFileExtraction(t *testing.T) {
	source := "::Sample::VALUE = 1\nself::OTHER = 2\n"
	path := filepath.Join(t.TempDir(), "constants.rb")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	batch, ok := ExtractRubyBatch([]RubyBatchFile{{Path: "constants.rb", SourcePath: path}})
	if !ok {
		t.Skip("RubyVM::AbstractSyntaxTree batch extraction is unavailable")
	}
	lines := splitLines(source)
	got := ExtractWithRubyBatch(batch, "constants.rb", "ruby", lines)
	want := Extract("constants.rb", "ruby", lines)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("batch symbols = %#v, want %#v", got, want)
	}
}

func TestParseRubyPrismDumpSymbols(t *testing.T) {
	source := `module Sample
  class Worker
    CONST = 1

    def self.build
    end

    def perform!
    end
  end
end
`
	dump := `@ ProgramNode (location: (1,0)-(11,3))
        +-- @ ModuleNode (location: (1,0)-(11,3))
            +-- @ ClassNode (location: (2,2)-(10,5))
                +-- @ ConstantWriteNode (location: (3,4)-(3,13))
                |   +-- name: :CONST
                |   +-- name_loc: (3,4)-(3,9) = "CONST"
                +-- @ DefNode (location: (5,4)-(6,7))
                |   +-- name: :build
                |   +-- name_loc: (5,13)-(5,18) = "build"
                |   +-- receiver:
                |   |   @ SelfNode (location: (5,8)-(5,12))
                |   +-- parameters: nil
                +-- @ DefNode (location: (8,4)-(9,7))
                    +-- name: :perform!
                    +-- name_loc: (8,8)-(8,16) = "perform!"
                    +-- receiver: nil
                    +-- parameters: nil
`
	symbols, ok := parseRubyPrismDump("worker.rb", "ruby", splitLines(source), dump)
	if !ok {
		t.Fatal("parseRubyPrismDump failed")
	}

	assertSymbol(t, symbols, "module", "Sample", 1)
	assertSymbol(t, symbols, "class", "Worker", 2)
	assertSymbol(t, symbols, "constant", "CONST", 3)
	assertSymbol(t, symbols, "method", "self.build", 5)
	assertSymbol(t, symbols, "method", "perform!", 8)
	assertSymbolEndLine(t, symbols, "module", "Sample", 11)
	assertSymbolEndLine(t, symbols, "class", "Worker", 10)
	assertSymbolEndLine(t, symbols, "method", "perform!", 9)
}

func TestParseRubyParseYDumpSymbols(t *testing.T) {
	source := `module Sample
  class Worker
    CONST = 1

    def self.build
    end

    def perform!
    end
  end
end
`
	dump := `# @ NODE_SCOPE (id: 1, line: 1, location: (1,0)-(11,3))
#     @ NODE_MODULE (id: 2, line: 1, location: (1,0)-(11,3))*
#           @ NODE_CLASS (id: 3, line: 2, location: (2,2)-(10,5))*
#                 @ NODE_CDECL (id: 4, line: 3, location: (3,4)-(3,13))*
#                 +- nd_vid: :CONST
#                 @ NODE_DEFS (id: 5, line: 5, location: (5,4)-(6,7))*
#                 +- nd_mid: :build
#                 @ NODE_DEFN (id: 6, line: 8, location: (8,4)-(9,7))*
#                 +- nd_mid: :perform!
`
	symbols, ok := parseRubyParseYDump("worker.rb", "ruby", splitLines(source), dump)
	if !ok {
		t.Fatal("parseRubyParseYDump failed")
	}

	assertSymbol(t, symbols, "module", "Sample", 1)
	assertSymbol(t, symbols, "class", "Worker", 2)
	assertSymbol(t, symbols, "constant", "CONST", 3)
	assertSymbol(t, symbols, "method", "self.build", 5)
	assertSymbol(t, symbols, "method", "perform!", 8)
	assertSymbolEndLine(t, symbols, "module", "Sample", 11)
	assertSymbolEndLine(t, symbols, "class", "Worker", 10)
	assertSymbolEndLine(t, symbols, "method", "perform!", 9)
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
