package symbols

import (
	"strings"

	treesitter "github.com/tree-sitter/go-tree-sitter"
	treesitterruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
)

type rubySymbolExtractor struct{}

type RubyExtractor struct {
	parser *treesitter.Parser
	query  *treesitter.Query
	cursor *treesitter.QueryCursor
}

const rubySymbolQuery = `
(class name: (_) @class.name) @class.definition
(module name: (_) @module.name) @module.definition
(method name: (_) @method.name) @method.definition
(singleton_method name: (_) @singleton_method.name) @singleton_method.definition
(assignment left: (constant) @constant.name) @constant.definition
`

func (rubySymbolExtractor) extract(path, language string, lines []string) ([]Symbol, bool) {
	extractor := NewRubyExtractor()
	if extractor == nil {
		return nil, false
	}
	defer extractor.Close()
	return extractor.extract(path, language, lines)
}

func NewRubyExtractor() *RubyExtractor {
	rubyLanguage := treesitter.NewLanguage(treesitterruby.Language())
	parser := treesitter.NewParser()
	if err := parser.SetLanguage(rubyLanguage); err != nil {
		parser.Close()
		return nil
	}
	query, err := treesitter.NewQuery(rubyLanguage, rubySymbolQuery)
	if err != nil {
		parser.Close()
		return nil
	}
	cursor := treesitter.NewQueryCursor()
	return &RubyExtractor{parser: parser, query: query, cursor: cursor}
}

func (extractor *RubyExtractor) Close() {
	if extractor == nil {
		return
	}
	extractor.cursor.Close()
	extractor.query.Close()
	extractor.parser.Close()
}

func ExtractWithRubyExtractor(extractor *RubyExtractor, path, language string, lines []string) []Symbol {
	if language != "ruby" || extractor == nil {
		return Extract(path, language, lines)
	}
	symbols, _ := extractor.extract(path, language, lines)
	return symbols
}

func (extractor *RubyExtractor) extract(path, language string, lines []string) ([]Symbol, bool) {
	source := []byte(strings.Join(lines, "\n"))
	tree := extractor.parser.Parse(source, nil)
	if tree == nil {
		return nil, false
	}
	defer tree.Close()

	var out []Symbol
	matches := extractor.cursor.Matches(extractor.query, tree.RootNode(), source)
	for match := matches.Next(); match != nil; match = matches.Next() {
		appendRubyMatch(match, extractor.query.CaptureNames(), path, language, source, lines, &out)
	}
	return out, true
}

func appendRubyMatch(match *treesitter.QueryMatch, captureNames []string, path, language string, source []byte, lines []string, out *[]Symbol) {
	var definitionNode, nameNode treesitter.Node
	definitionKind := ""
	for _, capture := range match.Captures {
		captureName := captureNames[capture.Index]
		if strings.HasSuffix(captureName, ".definition") {
			definitionNode = capture.Node
			definitionKind = strings.TrimSuffix(captureName, ".definition")
		} else if strings.HasSuffix(captureName, ".name") {
			nameNode = capture.Node
		}
	}
	if definitionKind == "" {
		return
	}

	kind := definitionKind
	name := ""
	switch definitionKind {
	case "class", "module":
		var ok bool
		name, ok = rubyConstantPath(&nameNode, source)
		if !ok {
			return
		}
	case "method":
		name = rubyMethodName(&nameNode, source)
	case "singleton_method":
		kind = "method"
		name = rubySingletonMethodName(&definitionNode, &nameNode, source)
	case "constant":
	default:
		return
	}
	appendRubySymbol(&definitionNode, &nameNode, kind, name, path, language, source, lines, out)
}

func appendRubySymbol(node, nameNode *treesitter.Node, kind, name, path, language string, source []byte, lines []string, out *[]Symbol) {
	if nameNode == nil || nameNode.EndByte() > uint(len(source)) {
		return
	}
	if name == "" {
		name = string(source[nameNode.StartByte():nameNode.EndByte()])
	}
	if name == "" {
		return
	}
	position := nameNode.StartPosition()
	line := int(position.Row) + 1
	symbol := buildSymbol(
		path,
		language,
		kind,
		name,
		line,
		int(position.Column)+1,
		sourceLine(lines, line),
		lines,
	)
	endLine := int(node.EndPosition().Row) + 1
	if endLine < symbol.Line {
		endLine = symbol.Line
	}
	symbol.EndLine = endLine
	*out = append(*out, symbol)
}

func rubyConstantPath(node *treesitter.Node, source []byte) (string, bool) {
	if node == nil || node.EndByte() > uint(len(source)) {
		return "", false
	}
	switch node.Kind() {
	case "constant":
		return string(source[node.StartByte():node.EndByte()]), true
	case "scope_resolution":
		scope := node.ChildByFieldName("scope")
		if scope != nil {
			if _, ok := rubyConstantPath(scope, source); !ok {
				return "", false
			}
		}
		name := strings.TrimPrefix(string(source[node.StartByte():node.EndByte()]), "::")
		return name, name != ""
	default:
		return "", false
	}
}

func rubyMethodName(nameNode *treesitter.Node, source []byte) string {
	if nameNode == nil || nameNode.EndByte() > uint(len(source)) {
		return ""
	}
	name := string(source[nameNode.StartByte():nameNode.EndByte()])
	if name == "~@" {
		return "~"
	}
	return name
}

func rubySingletonMethodName(node, nameNode *treesitter.Node, source []byte) string {
	name := rubyMethodName(nameNode, source)
	if name == "" || node.StartByte() > nameNode.StartByte() || nameNode.StartByte() > uint(len(source)) {
		return ""
	}
	prefix := strings.TrimSpace(string(source[node.StartByte():nameNode.StartByte()]))
	prefix = strings.TrimSpace(strings.TrimPrefix(prefix, "def"))
	if prefix == "" {
		return name
	}
	return prefix + name
}

func sourceLine(lines []string, line int) string {
	if line < 1 || line > len(lines) {
		return ""
	}
	return lines[line-1]
}
