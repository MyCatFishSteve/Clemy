package main

import "os"

// Configuration ...
type Config struct {
	DryRun  bool
	Verbose bool
}

// NewConfig will create a new config object based on current environment
// variables and return the pointer to the new object.
// Currently, it only specifies if DryRun is enabled but will include more parameters in the future
func NewConfiguration() *Config {
	_, DryRun := os.LookupEnv("CLEMY_DRY_RUN")
	_, Verbose := os.LookupEnv("CLEMY_VERBOSE")
	return &Config{
		DryRun:  DryRun,
		Verbose: Verbose,
	}
}
