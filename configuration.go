package main

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds runtime information of the program.
type Config struct {
	DryRun  bool // DryRun is used to disable destructive changes to AWS
	Verbose bool // Verbose is used to print additional information during runtime
	MaxAge  int  // Max age is used to filter instances based on their creation date
}

// NewConfiguration will create a new config object based on current environment
// variables and return the pointer to the new object.
// Currently, it only specifies if DryRun is enabled but will include more parameters in the future
func NewConfiguration() *Config {
	maxAge := 14
	_, DryRun := os.LookupEnv("CLEMY_DRY_RUN")
	_, Verbose := os.LookupEnv("CLEMY_VERBOSE")
	MaxAgeEnv, MaxAgeSet := os.LookupEnv("CLEMY_MAX_AGE")

	// If environment variable is not set correctly, the program should immediately exit.
	// Runtime should not be considered safe or abiding by the user wish.
	if MaxAgeSet {
		convertedAge, err := strconv.ParseInt(MaxAgeEnv, 0, 0)
		FatalError(err)
		maxAge = int(convertedAge)
	}

	// Print config information
	if DryRun {
		fmt.Println("Dry run enabled, program is running non-destructively")
	}

	return &Config{
		DryRun:  DryRun,
		Verbose: Verbose,
		MaxAge:  maxAge,
	}
}
