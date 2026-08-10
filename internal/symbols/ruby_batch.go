package symbols

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"time"
)

const rubyBatchTimeout = 10 * time.Minute

//go:embed ruby_batch.rb
var rubyBatchScript string

type RubyBatchFile struct {
	Path       string `json:"path"`
	SourcePath string `json:"source_path"`
}

type rubyBatchNode struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	EndLine  int    `json:"end_line"`
	Receiver bool   `json:"receiver"`
}

type rubyBatchResult struct {
	Path  string          `json:"path"`
	OK    bool            `json:"ok"`
	Nodes []rubyBatchNode `json:"nodes"`
}

type RubyBatch struct {
	results map[string]rubyBatchResult
}

func ExtractRubyBatch(files []RubyBatchFile) (*RubyBatch, bool) {
	if len(files) == 0 {
		return &RubyBatch{results: map[string]rubyBatchResult{}}, true
	}
	ruby, ok := rubyCommand()
	if !ok {
		return nil, false
	}

	var input bytes.Buffer
	encoder := json.NewEncoder(&input)
	for _, file := range files {
		if err := encoder.Encode(file); err != nil {
			return nil, false
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), rubyBatchTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, ruby, "-e", rubyBatchScript)
	cmd.Stdin = &input
	output, err := cmd.Output()
	if err != nil || ctx.Err() != nil {
		return nil, false
	}

	results := make(map[string]rubyBatchResult, len(files))
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var result rubyBatchResult
		if err := decoder.Decode(&result); err == io.EOF {
			break
		} else if err != nil || result.Path == "" {
			return nil, false
		}
		results[result.Path] = result
	}
	if len(results) != len(files) {
		return nil, false
	}
	return &RubyBatch{results: results}, true
}

func ExtractWithRubyBatch(batch *RubyBatch, path, language string, lines []string) []Symbol {
	if language != "ruby" || batch == nil {
		return Extract(path, language, lines)
	}
	result, ok := batch.results[path]
	if !ok || !result.OK {
		return Extract(path, language, lines)
	}
	return rubyBatchSymbols(path, language, lines, result.Nodes)
}

func rubyBatchSymbols(path, language string, lines []string, nodes []rubyBatchNode) []Symbol {
	var out []Symbol
	for _, node := range nodes {
		var symbol Symbol
		switch node.Type {
		case "MODULE", "CLASS":
			keyword := "module"
			kind := "module"
			if node.Type == "CLASS" {
				keyword = "class"
				kind = "class"
			}
			name, column := rubyClassOrModuleName(lines, node.Line, node.Column, keyword)
			if name == "" {
				continue
			}
			symbol = buildSymbol(path, language, kind, name, node.Line, column, sourceLine(lines, node.Line), lines)
		case "DEFN", "DEFS":
			if node.Name == "" {
				continue
			}
			column := rubyMethodNameColumn(lines, node.Name, node.Line, node.Column)
			name := node.Name
			if node.Receiver {
				column = rubyReceiverMethodNameColumn(lines, node.Name, node.Line, node.Column)
				name = rubyQualifiedMethodName(lines, name, node.Line, column)
			}
			if column == 0 {
				continue
			}
			symbol = buildSymbol(path, language, "method", name, node.Line, column, sourceLine(lines, node.Line), lines)
		case "CDECL":
			if node.Name == "" {
				continue
			}
			column := rubyNameColumn(lines, node.Name, node.Line, node.Column)
			symbol = buildSymbol(path, language, "constant", node.Name, node.Line, column, sourceLine(lines, node.Line), lines)
		default:
			continue
		}
		symbol.EndLine = node.EndLine
		out = append(out, symbol)
	}
	return out
}

func rubyMethodNameColumn(lines []string, name string, line, startColumn int) int {
	source := sourceLine(lines, line)
	if startColumn > len(source) {
		return 0
	}
	segment := source[startColumn:]
	defIndex := strings.Index(segment, "def")
	if defIndex < 0 {
		return 0
	}
	nameStart := startColumn + defIndex + len("def")
	index := strings.Index(source[nameStart:], name)
	if index < 0 {
		return 0
	}
	return nameStart + index + 1
}

func rubyReceiverMethodNameColumn(lines []string, name string, line, startColumn int) int {
	source := sourceLine(lines, line)
	if startColumn > len(source) {
		return 0
	}
	segment := source[startColumn:]
	index := strings.Index(segment, "."+name)
	separatorLength := 1
	if index < 0 {
		index = strings.Index(segment, "::"+name)
		separatorLength = 2
	}
	if index < 0 {
		return 0
	}
	return startColumn + index + separatorLength + 1
}
