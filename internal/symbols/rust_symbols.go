package symbols

import (
	"strings"

	treesitter "github.com/tree-sitter/go-tree-sitter"
	treesitterrust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
)

type rustSymbolExtractor struct{}

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
	source := []byte(strings.Join(lines, "\n"))
	parser := treesitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(treesitter.NewLanguage(treesitterrust.Language())); err != nil {
		return nil, false
	}
	tree := parser.Parse(source, nil)
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
