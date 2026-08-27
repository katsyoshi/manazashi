package symbols

import (
	"strings"
	"testing"
)

func TestExtractRustSymbolsUsesSyntaxTree(t *testing.T) {
	source := `mod models {
    pub struct Widget<T>
    where
        T: Clone,
    {
        value: T,
    }

    union Storage {
        integer: u64,
        float: f64,
    }

    enum State {
        Ready,
        Done,
    }

    trait Render {
        type Output;
        const ENABLED: bool;

        fn render(
            &self,
        ) -> Self::Output;

        fn default_render(&self) -> bool {
            true
        }
    }

    impl<T> Widget<T>
    where
        T: Clone,
    {
        pub fn new(value: T) -> Self {
            fn nested_helper() {}
            Self { value }
        }
    }
}

pub type WidgetId = u64;
pub const LIMIT: usize = 10;
pub static ENABLED: bool = true;

pub async fn load_widget(
    id: WidgetId,
) -> bool {
    id > 0
}

macro_rules! define_generated {
    () => {
        fn generated() {}
    };
}`

	symbols := Extract("src/lib.rs", "rust", splitLines(source))

	assertSymbol(t, symbols, "module", "models", 1)
	assertSymbol(t, symbols, "type", "Widget", 2)
	assertSymbolEndLine(t, symbols, "type", "Widget", 7)
	assertSymbol(t, symbols, "type", "Storage", 9)
	assertSymbol(t, symbols, "enum", "State", 14)
	assertSymbol(t, symbols, "trait", "Render", 19)
	assertSymbolEndLine(t, symbols, "trait", "Render", 30)
	assertSymbol(t, symbols, "type", "Output", 20)
	assertSymbol(t, symbols, "constant", "ENABLED", 21)
	assertSymbol(t, symbols, "method", "render", 23)
	assertSymbolEndLine(t, symbols, "method", "render", 25)
	assertSymbol(t, symbols, "method", "default_render", 27)
	assertSymbol(t, symbols, "method", "new", 36)
	assertSymbolEndLine(t, symbols, "method", "new", 39)
	assertSymbol(t, symbols, "function", "nested_helper", 37)
	assertSymbol(t, symbols, "type", "WidgetId", 43)
	assertSymbol(t, symbols, "constant", "LIMIT", 44)
	assertSymbol(t, symbols, "variable", "ENABLED", 45)
	assertSymbol(t, symbols, "function", "load_widget", 47)
	assertSymbolEndLine(t, symbols, "function", "load_widget", 51)
	assertSymbol(t, symbols, "macro", "define_generated", 53)
	assertSymbolEndLine(t, symbols, "macro", "define_generated", 57)
	assertNoSymbol(t, symbols, "function", "generated")

	load := findSymbol(t, symbols, "function", "load_widget")
	if load.Column != 14 {
		t.Fatalf("load_widget column = %d, want 14", load.Column)
	}
	if !strings.Contains(load.Signature, "pub async fn load_widget( id: WidgetId, ) -> bool") {
		t.Fatalf("load_widget signature = %q", load.Signature)
	}
}

func TestExtractRustSymbolsRecoversAroundSyntaxErrors(t *testing.T) {
	source := `fn before() {}

fn broken() {
    let value =
}

struct After {
    value: u64,
}`

	symbols := Extract("src/broken.rs", "rust", splitLines(source))

	assertSymbol(t, symbols, "function", "before", 1)
	assertSymbol(t, symbols, "function", "broken", 3)
	assertSymbol(t, symbols, "type", "After", 7)
}

func findSymbol(t *testing.T, symbols []Symbol, kind, name string) Symbol {
	t.Helper()
	for _, symbol := range symbols {
		if symbol.Kind == kind && symbol.Name == name {
			return symbol
		}
	}
	t.Fatalf("symbol %s %s not found in %+v", kind, name, symbols)
	return Symbol{}
}
