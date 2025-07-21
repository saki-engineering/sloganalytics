package slogger_test

import (
	"testing"

	"slogger"

	"github.com/gostaticanalysis/testutil"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := testutil.WithModules(t, analysistest.TestData(), nil)
	
	tests := []struct {
		name    string
		pkgPath string
	}{
		{
			name:    "missing WithAttrs method",
			pkgPath: "missing_withattrs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysistest.Run(t, testdata, slogger.Analyzer, tt.pkgPath)
		})
	}
}
