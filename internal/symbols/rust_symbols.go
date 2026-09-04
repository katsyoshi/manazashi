package symbols

import (
	"strings"

	treesitter "github.com/tree-sitter/go-tree-sitter"
	treesitterrust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
)

type rustSymbolExtractor struct{}

// RustExtractor owns the parser used to extract Rust symbols. It is
// safe to reuse sequentially across files, as done by rebuild and update.
type RustExtractor struct {
	parser *treesitter.Parser
}

var rustSymbolKinds = map[string]string{
	"associated_type":  "type",
	"const_item":       "constant",
	"enum_item":        "enum",
	"macro_definition": "macro",
	"mod_item":         "module",
	"static_item":      "variable",
	"struct_item":      "type",
	"trait_item":       "trait",
	"type_item":        "type",
	"union_item":       "type",
}

func (rustSymbolExtractor) extract(path, language string, lines []string) ([]Symbol, bool) {
	extractor := NewRustExtractor()
	if extractor == nil {
		return nil, false
	}
	defer extractor.Close()
	return extractor.extract(path, language, lines)
}

func NewRustExtractor() *RustExtractor {
	rustLanguage := treesitter.NewLanguage(treesitterrust.Language())
	parser := treesitter.NewParser()
	if err := parser.SetLanguage(rustLanguage); err != nil {
		parser.Close()
		return nil
	}
	return &RustExtractor{parser: parser}
}

func (extractor *RustExtractor) Close() {
	if extractor == nil {
		return
	}
	extractor.parser.Close()
}

func ExtractWithRustExtractor(extractor *RustExtractor, path, language string, lines []string) []Symbol {
	if language != "rust" || extractor == nil {
		return Extract(path, language, lines)
	}
	symbols, _ := extractor.extract(path, language, lines)
	return symbols
}

func (extractor *RustExtractor) extract(path, language string, lines []string) ([]Symbol, bool) {
	source := []byte(strings.Join(lines, "\n"))
	tree := extractor.parser.Parse(source, nil)
	if tree == nil {
		return nil, false
	}
	defer tree.Close()

	var out []Symbol
	walkRustSymbols(tree.RootNode(), false, path, language, source, lines, &out)
	return out, true
}

func walkRustSymbols(node *treesitter.Node, methodOwner bool, path, language string, source []byte, lines []string, out *[]Symbol) {
	kind := node.Kind()
	childMethodOwner := methodOwner
	switch kind {
	case "impl_item":
		childMethodOwner = true
	case "trait_item":
		appendRustSymbol(node, rustSymbolKinds[kind], path, language, source, lines, out)
		childMethodOwner = true
	case "function_item", "function_signature_item":
		symbolKind := "function"
		if methodOwner {
			symbolKind = "method"
		}
		appendRustSymbol(node, symbolKind, path, language, source, lines, out)
		childMethodOwner = false
	default:
		if symbolKind, ok := rustSymbolKinds[kind]; ok {
			appendRustSymbol(node, symbolKind, path, language, source, lines, out)
		}
	}

	for i := uint(0); i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child != nil {
			walkRustSymbols(child, childMethodOwner, path, language, source, lines, out)
		}
	}
}

func appendRustSymbol(node *treesitter.Node, kind, path, language string, source []byte, lines []string, out *[]Symbol) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil || nameNode.EndByte() > uint(len(source)) {
		return
	}
	name := string(source[nameNode.StartByte():nameNode.EndByte()])
	if name == "" {
		return
	}
	position := nameNode.StartPosition()
	symbol := buildSymbol(
		path,
		language,
		kind,
		name,
		int(position.Row)+1,
		int(position.Column)+1,
		rustSignature(node, source),
		lines,
	)
	endLine := int(node.EndPosition().Row) + 1
	if endLine < symbol.Line {
		endLine = symbol.Line
	}
	symbol.EndLine = endLine
	*out = append(*out, symbol)
}

func rustSignature(node *treesitter.Node, source []byte) string {
	start, end := node.StartByte(), node.EndByte()
	if body := node.ChildByFieldName("body"); body != nil {
		end = body.StartByte()
	}
	if start > end || end > uint(len(source)) {
		return ""
	}
	return strings.Join(strings.Fields(string(source[start:end])), " ")
}
