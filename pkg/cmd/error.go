package cmd

import (
	"fmt"
	"os"
)

// ErrExitCode is the exit code used for erorrs writtem tp stderr
const ErrExitCode = 1

// CheckErrExit checks if the error is not nil, if not the error is written to stderr and the exits
// the application with an error exit code
func CheckErrExit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ErrExitCode)
	}
}
