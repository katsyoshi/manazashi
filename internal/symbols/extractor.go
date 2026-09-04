package symbols

type reusableExtractor interface {
	extract(path, language string, lines []string) ([]Symbol, bool)
	Close()
}

// Extractor owns reusable parser-backed symbol extractors. It is intended for
// sequential use by a build operation; callers must not use it concurrently.
type Extractor struct {
	extractors map[string]reusableExtractor
}

// NewExtractor initializes the parser-backed extractors used by the indexer.
func NewExtractor() *Extractor {
	extractor := &Extractor{extractors: map[string]reusableExtractor{}}
	ruby := NewRubyExtractor()
	if ruby != nil {
		extractor.extractors["ruby"] = ruby
	}
	java := NewJavaExtractor()
	if java != nil {
		extractor.extractors["java"] = java
	}
	rust := NewRustExtractor()
	if rust != nil {
		extractor.extractors["rust"] = rust
	}
	return extractor
}

// Close releases all parser-backed extractor resources.
func (extractor *Extractor) Close() {
	if extractor == nil {
		return
	}
	for _, reusable := range extractor.extractors {
		reusable.Close()
	}
}

// ExtractWithExtractor uses reusable parser-backed extraction where available
// and preserves Extract's fallback behavior for all other languages.
func ExtractWithExtractor(extractor *Extractor, path, language string, lines []string) []Symbol {
	if extractor == nil {
		return Extract(path, language, lines)
	}
	if reusable, ok := extractor.extractors[language]; ok {
		symbols, ok := reusable.extract(path, language, lines)
		if ok {
			return symbols
		}
	}
	return Extract(path, language, lines)
}
