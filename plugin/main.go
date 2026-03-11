package main

import (
	"github.com/mitchellh/mapstructure"
	"github.com/sabakaru/loglinter/pkg/analyzer"
	"golang.org/x/tools/go/analysis"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	var config analyzer.Config
	if err := mapstructure.Decode(conf, &config); err != nil {
		return nil, err
	}

	if len(config.SensitiveWords) > 0 {
		analyzer.SetSensitiveWords(config.SensitiveWords)
	}

	return []*analysis.Analyzer{analyzer.Analyzer}, nil
}
