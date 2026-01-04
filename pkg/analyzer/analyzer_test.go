package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestInterruptionAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), InterruptCheckAnalyzer, "./...")
}
