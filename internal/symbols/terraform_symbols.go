package symbols

import (
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type terraformSymbolExtractor struct{}

type terraformBlockSpec struct {
	kind      string
	labels    []string
	nameLabel int
}

var terraformBlockSpecs = map[string]terraformBlockSpec{
	"resource": {kind: "resource", labels: []string{"type", "name"}, nameLabel: 1},
	"data":     {kind: "data", labels: []string{"type", "name"}, nameLabel: 1},
	"module":   {kind: "module", labels: []string{"name"}},
	"variable": {kind: "variable", labels: []string{"name"}},
	"output":   {kind: "output", labels: []string{"name"}},
	"provider": {kind: "provider", labels: []string{"name"}},
	"check":    {kind: "check", labels: []string{"name"}},
}

func (terraformSymbolExtractor) extract(path, language string, lines []string) ([]Symbol, bool) {
	source := []byte(strings.Join(lines, "\n"))
	parser := hclparse.NewParser()
	if strings.HasSuffix(strings.ToLower(path), ".json") {
		file, _ := parser.ParseJSON(source, path)
		return extractTerraformJSONSymbols(file, path, language, lines), true
	}

	file, _ := parser.ParseHCL(source, path)
	return extractTerraformNativeSymbols(file, path, language, lines), true
}

func extractTerraformNativeSymbols(file *hcl.File, path, language string, lines []string) []Symbol {
	if file == nil {
		return nil
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	out := make([]Symbol, 0, len(body.Blocks))
	for _, block := range body.Blocks {
		spec, ok := terraformBlockSpecs[block.Type]
		if !ok || len(block.Labels) != len(spec.labels) || len(block.LabelRanges) <= spec.nameLabel {
			continue
		}
		labelRange := block.LabelRanges[spec.nameLabel]
		symbol := terraformBlockSymbol(block.Type, block.Labels, spec, labelRange, path, language, lines)
		if block.CloseBraceRange.End.Line >= symbol.Line {
			symbol.EndLine = block.CloseBraceRange.End.Line
		}
		out = append(out, symbol)
	}
	return out
}

func extractTerraformJSONSymbols(file *hcl.File, path, language string, lines []string) []Symbol {
	if file == nil {
		return nil
	}
	schema := &hcl.BodySchema{Blocks: make([]hcl.BlockHeaderSchema, 0, len(terraformBlockSpecs))}
	for blockType, spec := range terraformBlockSpecs {
		schema.Blocks = append(schema.Blocks, hcl.BlockHeaderSchema{Type: blockType, LabelNames: spec.labels})
	}
	content, _, _ := file.Body.PartialContent(schema)
	if content == nil {
		return nil
	}

	out := make([]Symbol, 0, len(content.Blocks))
	for _, block := range content.Blocks {
		spec := terraformBlockSpecs[block.Type]
		if len(block.Labels) != len(spec.labels) || len(block.LabelRanges) <= spec.nameLabel {
			continue
		}
		symbol := terraformBlockSymbol(block.Type, block.Labels, spec, block.LabelRanges[spec.nameLabel], path, language, lines)
		if endLine := block.Body.MissingItemRange().Start.Line; endLine >= symbol.Line {
			symbol.EndLine = endLine
		}
		out = append(out, symbol)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Line == out[j].Line {
			return out[i].Column < out[j].Column
		}
		return out[i].Line < out[j].Line
	})
	return out
}

func terraformBlockSymbol(blockType string, labels []string, spec terraformBlockSpec, labelRange hcl.Range, path, language string, lines []string) Symbol {
	column := labelRange.Start.Column
	if column > 0 {
		column++ // HCL label ranges include the opening quote.
	}
	symbol := buildSymbol(
		path,
		language,
		spec.kind,
		labels[spec.nameLabel],
		labelRange.Start.Line,
		column,
		terraformBlockSignature(blockType, labels),
		lines,
	)
	return symbol
}

func terraformBlockSignature(blockType string, labels []string) string {
	parts := make([]string, 1, len(labels)+1)
	parts[0] = blockType
	for _, label := range labels {
		parts = append(parts, strconv.Quote(label))
	}
	return strings.Join(parts, " ")
}
