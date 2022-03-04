package goimports

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"math"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "goimports-custom is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "goimports-custom",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func extractPath(spec *ast.ImportSpec) string {
	path, err := strconv.Unquote(spec.Path.Value)
	if err == nil {
		return path
	}
	return strings.Trim(spec.Path.Value, "`")
}

func newText(fset *token.FileSet, groups [][]*ast.ImportSpec) []byte {
	if len(groups) == 0 {
		return []byte("")
	}

	var buffer bytes.Buffer
	buffer.WriteString("import (\n")
	for _, spec := range groups[0] {
		format.Node(&buffer, fset, spec)
		buffer.WriteString("\n")
	}
	for _, group := range groups[1:] {
		buffer.WriteString("\n")
		for _, spec := range group {
			format.Node(&buffer, fset, spec)
			buffer.WriteString("\n")
		}
	}
	buffer.WriteString(")\n")
	return buffer.Bytes()
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if len(sections) == 0 {
		sections = []Section{
			{Type: Standard},
			{Type: Default},
		}
	}
	var defaultIncluded bool
	for _, section := range sections {
		if section.Type == Default {
			defaultIncluded = true
		}
	}
	if !defaultIncluded {
		sections = append(sections, Section{Type: Default})
	}

	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	var start token.Pos = math.MaxInt
	var imports []*ast.GenDecl
	var imported []*ast.ImportSpec
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		decl, ok := n.(*ast.GenDecl)
		if !ok {
			return
		}
		if pos := decl.Pos(); pos < start {
			start = pos
		}
		if decl.Tok != token.IMPORT {
			return
		}
		imports = append(imports, decl)
		for _, spec := range decl.Specs {
			imported = append(imported, spec.(*ast.ImportSpec))
		}
	})

	groups := make([][]*ast.ImportSpec, len(sections))
	defaultIndex := sections.DefaultIndex()
loop:
	for _, spec := range imported {
		path := extractPath(spec)
		for i, section := range sections {
			if section.Match(path) {
				groups[i] = append(groups[i], spec)
				continue loop
			}
		}
		groups[defaultIndex] = append(groups[defaultIndex], spec)
	}

	for _, group := range groups {
		sort.SliceStable(group, func(i, j int) bool { return extractPath(group[i]) < extractPath(group[j]) })
	}

	for _, decl := range imports {
		pass.Report(analysis.Diagnostic{
			Pos:      decl.Pos(),
			End:      decl.End(),
			Category: "style",
			Message:  "",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "",
				TextEdits: []analysis.TextEdit{{
					Pos: decl.Pos(),
					End: decl.End(),
				}},
			}},
		})
	}
	pass.Report(analysis.Diagnostic{
		Pos:      start,
		End:      0,
		Category: "style",
		Message:  "",
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "",
			TextEdits: []analysis.TextEdit{{
				Pos:     start,
				End:     start,
				NewText: newText(pass.Fset, groups),
			}},
		}},
	})

	return nil, nil
}
