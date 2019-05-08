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

	DryRun := false
	Verbose := false
	MaxAge := 28

	DryRunEnv, _ := os.LookupEnv("CLEMY_DRY_RUN")
	VerboseEnv, _ := os.LookupEnv("CLEMY_VERBOSE")
	MaxAgeEnv, _ := os.LookupEnv("CLEMY_MAX_AGE")

	if len(DryRunEnv) > 0 {
		DryRun = true
	}

	if len(VerboseEnv) > 0 {
		Verbose = true
	}

	// If environment variable is not set correctly, the program should immediately exit.
	// Runtime should not be considered safe or abiding by the user wish.
	if len(MaxAgeEnv) > 0 {
		maxAge64, err := strconv.ParseInt(MaxAgeEnv, 0, 0)
		FatalError(err)
		MaxAge = int(maxAge64)
	}

	// Print config information
	if DryRun {
		fmt.Println("Dry run enabled, program is running non-destructively")
	}

	if Verbose {
		fmt.Println("Verbose run enabled, program will print additional information")
	}

	return &Config{
		DryRun:  DryRun,
		Verbose: Verbose,
		MaxAge:  MaxAge,
	}
}
