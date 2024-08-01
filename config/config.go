package config

import (
	"flag"
)

// NOTE: Variables in this file are initialized at the start of the program
// They should be treated as constants and never modified

// The log level to use (debug, info, warn, error)
// Default: info
var LogLevel string

// The path to the directory containing filter files
// Default: ./filters
var FiltersDir string

func LoadConfig() {
	flag.StringVar(&LogLevel, "loglevel", "info", "The log level to use (debug, info, warn, error)")
	flag.StringVar(&FiltersDir, "filters", "./filters", "The path to the directory containing filter files")

	flag.Parse()
}
