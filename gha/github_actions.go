// Package gha provides functions to interact with the GitHub Actions runtime.
package gha

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sethvargo/go-githubactions"
)

// InGitHubActions indicates whether this application is being run within GitHub Actions.
func InGitHubActions() bool {
	inGitHubActions, _ := strconv.ParseBool(os.Getenv("GITHUB_ACTIONS"))
	return inGitHubActions
}

// ReadInput gets the input by the given name.
func ReadInput(name string) (value string, err error) {
	value = githubactions.GetInput(name)
	if value == "" {
		err = fmt.Errorf("input not set")
	}
	return
}

// WriteOutput writes an output parameter.
func WriteOutput(name, value string) {
	githubactions.SetOutput(name, value)
}

// AddMask requests GitHub Actions to mask the given value in the logs.
func AddMask(value string) {
	githubactions.AddMask(value)
}
