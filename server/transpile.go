package server

import (
	"github.com/scriptmaster/openagent/server/transpile"
)

// ============================================================================
// MAIN TRANSPILE INTERFACE
// ============================================================================
// This file provides the main interface to the transpilation system
// All actual transpilation logic has been moved to the transpile package
// ============================================================================
// TranspileHtmlToTsx delegates to the transpile package
func TranspileHtmlToTsx(inputPath, outputPath string) error {
	return transpile.TranspileHtmlToTsx(inputPath, outputPath)
}

// TranspileLayoutToTsx delegates to the transpile package
func TranspileLayoutToTsx(inputPath, outputPath string) error {
	return transpile.TranspileLayoutToTsx(inputPath, outputPath)
}

// TranspileAllTemplates delegates to the transpile package
func TranspileAllTemplates() error {
	return transpile.TranspileAllTemplates()
}

// TSX2JS delegates to the transpile package
func TSX2JS(tsxStr string) string {
	return transpile.TSX2JS(tsxStr)
}
