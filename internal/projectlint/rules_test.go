package projectlint_test

import (
	"testing"

	"github.com/TradeLayers/BE/internal/projectlint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestIntNameAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), projectlint.IntNameAnalyzer, "intname")
}

func TestExplicitInitAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), projectlint.ExplicitInitAnalyzer, "explicitinit")
}
