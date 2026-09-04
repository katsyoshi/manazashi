package symbols

import (
	"strings"

	treesitter "github.com/tree-sitter/go-tree-sitter"
	treesitterjava "github.com/tree-sitter/tree-sitter-java/bindings/go"
)

type javaSymbolExtractor struct{}

// JavaExtractor owns the parser, query, and cursor used to extract Java
// symbols. It is safe to reuse sequentially across files.
type JavaExtractor struct {
	parser *treesitter.Parser
	query  *treesitter.Query
	cursor *treesitter.QueryCursor
}

var javaSymbolKinds = map[string]string{
	"annotation_type_declaration":     "interface",
	"class_declaration":               "class",
	"constructor_declaration":         "method",
	"compact_constructor_declaration": "method",
	"enum_declaration":                "enum",
	"interface_declaration":           "interface",
	"method_declaration":              "method",
	"record_declaration":              "class",
}

const javaSymbolQuery = `
[
  (class_declaration)
  (interface_declaration)
  (enum_declaration)
  (record_declaration)
  (annotation_type_declaration)
  (method_declaration)
  (constructor_declaration)
  (compact_constructor_declaration)
] @definition
`

func (javaSymbolExtractor) extract(path, language string, lines []string) ([]Symbol, bool) {
	extractor := NewJavaExtractor()
	if extractor == nil {
		return nil, false
	}
	defer extractor.Close()
	return extractor.extract(path, language, lines)
}

func NewJavaExtractor() *JavaExtractor {
	javaLanguage := treesitter.NewLanguage(treesitterjava.Language())
	parser := treesitter.NewParser()
	if err := parser.SetLanguage(javaLanguage); err != nil {
		parser.Close()
		return nil
	}
	query, err := treesitter.NewQuery(javaLanguage, javaSymbolQuery)
	if err != nil {
		parser.Close()
		return nil
	}
	return &JavaExtractor{parser: parser, query: query, cursor: treesitter.NewQueryCursor()}
}

func (extractor *JavaExtractor) Close() {
	if extractor == nil {
		return
	}
	extractor.cursor.Close()
	extractor.query.Close()
	extractor.parser.Close()
}

func ExtractWithJavaExtractor(extractor *JavaExtractor, path, language string, lines []string) []Symbol {
	if language != "java" || extractor == nil {
		return Extract(path, language, lines)
	}
	symbols, _ := extractor.extract(path, language, lines)
	return symbols
}

func (extractor *JavaExtractor) extract(path, language string, lines []string) ([]Symbol, bool) {
	source := []byte(strings.Join(lines, "\n"))
	tree := extractor.parser.Parse(source, nil)
	if tree == nil {
		return nil, false
	}
	defer tree.Close()

	var out []Symbol
	captures := extractor.cursor.Captures(extractor.query, tree.RootNode(), source)
	for match, captureIndex := captures.Next(); match != nil; match, captureIndex = captures.Next() {
		capture := match.Captures[captureIndex]
		node := &capture.Node
		kind, ok := javaSymbolKinds[node.Kind()]
		if !ok {
			continue
		}
		appendJavaSymbol(node, kind, path, language, source, lines, &out)
	}
	return out, true
}

func appendJavaSymbol(node *treesitter.Node, kind, path, language string, source []byte, lines []string, out *[]Symbol) {
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
		javaSignature(node, source),
		lines,
	)
	endLine := int(node.EndPosition().Row) + 1
	if endLine < symbol.Line {
		endLine = symbol.Line
	}
	symbol.EndLine = endLine
	*out = append(*out, symbol)
}

func javaSignature(node *treesitter.Node, source []byte) string {
	start, end := node.StartByte(), node.EndByte()
	if body := node.ChildByFieldName("body"); body != nil {
		end = body.StartByte()
	}
	if start > end || end > uint(len(source)) {
		return ""
	}
	return strings.Join(strings.Fields(string(source[start:end])), " ")
}
