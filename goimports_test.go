package goimports_test

import (
	"testing"

	"github.com/Warashi/goimports-custom"

	"github.com/gostaticanalysis/testutil"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := testutil.WithModules(t, analysistest.TestData(), nil)
	analyzer := goimports.Analyzer
	analyzer.Flags.Set("section", "Standard")
	analyzer.Flags.Set("section", "Default")
	analyzer.Flags.Set("section", "Prefix(a)")
	analysistest.RunWithSuggestedFixes(t, testdata, analyzer, "a")
}
