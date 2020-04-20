package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	gha "github.com/sethvargo/go-githubactions"
)

func readInput() input {
	isGitHubActions := os.Getenv("GITHUB_ACTIONS")

	if isGitHubActions == "true" {
		return readInputGitHubActions()
	}
	return readInputLocal()
}

// Read input from the GitHub Actions run environment
func readInputGitHubActions() input {
	token := getInputOrExit("token")
	gha.AddMask(token)

	organization := getInputOrExit("organization")
	workspace := getInputOrExit("workspace")
	message := getInputOrExit("message")
	directory := getInputOrExit("directory")

	speculative, err := strconv.ParseBool(getInputOrExit("speculative"))
	if err != nil {
		exitErrorf("Could not parse input 'speculative' as bool: %v\n", err)
	}

	return input{
		Token:        token,
		Organization: organization,
		Workspace:    workspace,
		Message:      message,
		Directory:    directory,
		Speculative:  speculative,
	}
}

func getInputOrExit(identifier string) string {
	val := gha.GetInput(identifier)
	if val == "" {
		exitErrorf("Missing input '%v'\n", identifier)
	}
	return val
}

// Read input from a local development environment
func readInputLocal() input {
	bytes, err := ioutil.ReadFile("input.json")
	if err != nil {
		exitErrorf("Could not read 'input.json': %v\n", err)
	}

	var in input
	err = json.Unmarshal(bytes, &in)
	if err != nil {
		exitErrorf("Could not parse 'input.json': %v\n", err)
	}

	return in
}
