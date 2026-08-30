package symbols

import (
	"strings"
	"testing"
)

func TestExtractRubySymbolsUsesSyntaxTree(t *testing.T) {
	source := `module Sample
  class Admin::Worker < Base
    CONST = 1
    Sample::QUALIFIED = 2

    def self.build(...)
      new(...)
    end

    def quoter.quote(value)
    end

    def perform!(value = nil)
      value
    end

    def []=(key, value) = value
  end
end
`
	symbols := Extract("worker.rb", "ruby", splitLines(source))

	assertSymbol(t, symbols, "module", "Sample", 1)
	assertSymbolEndLine(t, symbols, "module", "Sample", 19)
	assertSymbol(t, symbols, "class", "Admin::Worker", 2)
	assertSymbolEndLine(t, symbols, "class", "Admin::Worker", 18)
	assertSymbol(t, symbols, "constant", "CONST", 3)
	assertNoSymbol(t, symbols, "constant", "QUALIFIED")
	assertSymbol(t, symbols, "method", "self.build", 6)
	assertSymbolEndLine(t, symbols, "method", "self.build", 8)
	assertSymbol(t, symbols, "method", "quoter.quote", 10)
	assertSymbol(t, symbols, "method", "perform!", 13)
	assertSymbolEndLine(t, symbols, "method", "perform!", 15)
	assertSymbol(t, symbols, "method", "[]=", 17)

	build := findSymbol(t, symbols, "method", "self.build")
	if build.Column != 14 {
		t.Fatalf("self.build column = %d, want 14", build.Column)
	}
	if !strings.Contains(build.Signature, "def self.build(...)") {
		t.Fatalf("self.build signature = %q", build.Signature)
	}
}

func TestExtractRubySymbolsRecoversAroundSyntaxErrors(t *testing.T) {
	source := `def before
end

def broken
  value =
end

class After
end
`
	symbols := Extract("broken.rb", "ruby", splitLines(source))

	assertSymbol(t, symbols, "method", "before", 1)
	assertSymbol(t, symbols, "method", "broken", 4)
	assertSymbol(t, symbols, "class", "After", 8)
}

func TestExtractRubySymbolsNormalizesStoredNames(t *testing.T) {
	source := `class ::Rooted
end

class self.class::Dynamic
end

def (instance.foo).bar
end

def ~@
end
`
	symbols := Extract("names.rb", "ruby", splitLines(source))

	assertSymbol(t, symbols, "class", "Rooted", 1)
	assertNoSymbol(t, symbols, "class", "::Rooted")
	assertNoSymbol(t, symbols, "class", "self.class::Dynamic")
	assertSymbol(t, symbols, "method", "(instance.foo).bar", 7)
	assertSymbol(t, symbols, "method", "~", 10)
}
