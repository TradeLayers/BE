package main

import (
	"github.com/TradeLayers/BE/internal/projectlint"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		projectlint.IntNameAnalyzer,
		projectlint.ExplicitInitAnalyzer,
	)
}
