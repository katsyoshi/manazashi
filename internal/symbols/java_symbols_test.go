package symbols

import "testing"

func TestExtractJavaSymbols(t *testing.T) {
	source := `// class CommentedOut {}
public class Outer {
    private String field = "class StringLiteral {}";

    @Override
    public Outer() {}

    public Outer(int value) {
        this.field = "method fake() {}";
    }

    @Override
    public void run() {}

    static class Inner {
        void nested() {}
    }

    interface Nested {
        void contract();
    }

    enum Kind {
        ONE;
        void enumMethod() {}
    }
}

record Point(int x, int y) {
    Point {
        if (x < 0) throw new IllegalArgumentException();
    }

    Point(int x, int y, String label) {}
}

interface Contract {
    void execute();
}

enum Top { A }

@interface Marker {
    String value();
}`

	symbols := Extract("Example.java", "java", splitLines(source))
	wantNames := []string{
		"Outer", "Outer", "Outer", "run", "Inner", "nested", "Nested", "contract",
		"Kind", "enumMethod", "Point", "Point", "Point", "Contract", "execute", "Top", "Marker",
	}
	if len(symbols) != len(wantNames) {
		t.Fatalf("symbol count = %d, want %d: %+v", len(symbols), len(wantNames), symbols)
	}
	for i, want := range wantNames {
		if symbols[i].Name != want {
			t.Fatalf("symbol %d = %q, want %q", i, symbols[i].Name, want)
		}
	}

	assertJavaSymbol(t, symbols, "class", "Outer", 2)
	assertJavaSymbol(t, symbols, "method", "run", 13)
	assertJavaSymbol(t, symbols, "class", "Inner", 15)
	assertJavaSymbol(t, symbols, "interface", "Nested", 19)
	assertJavaSymbol(t, symbols, "enum", "Kind", 23)
	assertJavaSymbol(t, symbols, "class", "Point", 29)
	assertJavaSymbol(t, symbols, "interface", "Contract", 37)
	assertJavaSymbol(t, symbols, "enum", "Top", 41)
	assertJavaSymbol(t, symbols, "interface", "Marker", 43)
	assertJavaEndLine(t, symbols, "class", "Outer", 27)
	assertJavaEndLine(t, symbols, "class", "Point", 35)
	assertJavaNoSymbol(t, symbols, "CommentedOut")
	assertJavaNoSymbol(t, symbols, "StringLiteral")
	assertJavaNoSymbol(t, symbols, "fake")

	for _, symbol := range symbols {
		if symbol.Name == "value" {
			t.Fatalf("annotation element was indexed as a symbol: %+v", symbol)
		}
	}
	constructor := symbols[1]
	if constructor.Kind != "method" || constructor.Column != 12 {
		t.Fatalf("constructor = %+v, want method at column 12", constructor)
	}
	if constructor.Signature != "@Override public Outer()" {
		t.Fatalf("constructor signature = %q", constructor.Signature)
	}
	compact := symbols[11]
	if compact.Signature != "Point" {
		t.Fatalf("compact constructor signature = %q", compact.Signature)
	}
}

func TestExtractJavaSymbolsRecoversAroundSyntaxErrors(t *testing.T) {
	source := `class Before {
    void okay() {}
}

class Broken {
    void broken( {
}

class After {
    void recovered() {}
}`

	symbols := Extract("Broken.java", "java", splitLines(source))
	assertJavaSymbol(t, symbols, "class", "Before", 1)
	assertJavaSymbol(t, symbols, "method", "okay", 2)
	assertJavaSymbol(t, symbols, "class", "Broken", 5)
	assertJavaSymbol(t, symbols, "class", "After", 9)
	assertJavaSymbol(t, symbols, "method", "recovered", 10)
}

func TestExtractJavaSymbolsReusesExtractor(t *testing.T) {
	extractor := NewJavaExtractor()
	if extractor == nil {
		t.Fatal("NewJavaExtractor returned nil")
	}
	defer extractor.Close()

	sources := []string{
		"class First { void one() {} }",
		"class Second { void two() {} void two(int value) {} }",
	}
	for _, source := range sources {
		lines := splitLines(source)
		want := Extract("Example.java", "java", lines)
		got := ExtractWithJavaExtractor(extractor, "Example.java", "java", lines)
		if len(got) != len(want) {
			t.Fatalf("reused extractor returned %d symbols, want %d: got=%+v want=%+v", len(got), len(want), got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("reused extractor symbol %d = %+v, want %+v", i, got[i], want[i])
			}
		}
	}
}

func assertJavaSymbol(t *testing.T, symbols []Symbol, kind, name string, line int) {
	t.Helper()
	for _, symbol := range symbols {
		if symbol.Kind == kind && symbol.Name == name && symbol.Line == line {
			return
		}
	}
	t.Fatalf("symbol %s %s at line %d not found in %+v", kind, name, line, symbols)
}

func assertJavaEndLine(t *testing.T, symbols []Symbol, kind, name string, endLine int) {
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

func assertJavaNoSymbol(t *testing.T, symbols []Symbol, name string) {
	t.Helper()
	for _, symbol := range symbols {
		if symbol.Name == name {
			t.Fatalf("unexpected symbol %s: %+v", name, symbol)
		}
	}
}
