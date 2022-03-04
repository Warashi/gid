package goimports

import (
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/k0kubun/pp/v3"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "gid is deterministic goimports"

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "gid",
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

func text(fset *token.FileSet, node interface{}) string {
	var builder strings.Builder
	printer.Fprint(&builder, fset, node)
	return builder.String()

}

func newText(fset *token.FileSet, groups [][]*ast.ImportSpec) string {
	if len(groups) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("import (\n")
	for _, spec := range groups[0] {
		builder.WriteString("\t")
		format.Node(&builder, fset, spec)
		builder.WriteString("\n")
	}
	for _, group := range groups[1:] {
		builder.WriteString("\n")
		for _, spec := range group {
			builder.WriteString("\t")
			format.Node(&builder, fset, spec)
			builder.WriteString("\n")
		}
	}
	builder.WriteString(")")
	return builder.String()
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
	if len(imports) == 0 {
		return nil, nil
	}
	uniq := make(map[string]*ast.ImportSpec, len(imports))
	for _, spec := range imported {
		uniq[text(pass.Fset, spec)] = spec
	}
	imported = make([]*ast.ImportSpec, 0, len(uniq))
	for _, spec := range uniq {
		imported = append(imported, spec)
	}

	sort.SliceStable(imports, func(i, j int) bool { return imports[i].Pos() < imports[j].Pos() })
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

	applied := newText(pass.Fset, groups)
	if len(imports) == 1 && applied == text(pass.Fset, imports[0]) {
		return nil, nil
	}
	log.Println(cmp.Diff(text(pass.Fset, imports[0]), applied))
	decl := imports[0]
	pass.Report(analysis.Diagnostic{
		Pos:      decl.Pos(),
		End:      decl.End(),
		Category: "style",
		Message:  "not gid'ed",
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "apply gdi",
			TextEdits: []analysis.TextEdit{{
				Pos:     decl.Pos(),
				End:     decl.End(),
				NewText: []byte(applied),
			}},
		}},
	})

	for _, decl := range imports[1:] {
		pass.Report(analysis.Diagnostic{
			Pos:      decl.Pos(),
			End:      decl.End(),
			Category: "style",
			Message:  "not gid'ed",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "apply gid",
				TextEdits: []analysis.TextEdit{{
					Pos: decl.Pos(),
					End: decl.End(),
				}},
			}},
		})
	}

	return nil, nil
}
