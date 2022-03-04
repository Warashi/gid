package gid_test

import (
	"testing"

	"github.com/Warashi/gid"

	"github.com/gostaticanalysis/testutil"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := testutil.WithModules(t, analysistest.TestData(), nil)
	analyzer := gid.Analyzer
	analyzer.Flags.Set("section", "Standard")
	analyzer.Flags.Set("section", "Default")
	analyzer.Flags.Set("section", "Prefix(a)")
	analysistest.RunWithSuggestedFixes(t, testdata, analyzer, "a")
}
