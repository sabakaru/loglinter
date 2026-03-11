package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/sabakaru/loglinter/pkg/analyzer"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
