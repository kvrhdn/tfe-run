package main

import (
	"strconv"

	gha "github.com/sethvargo/go-githubactions"
)

func readInput() input {
	token := getInputOrFail("token")
	gha.AddMask(token)

	organization := getInputOrFail("organization")
	workspace := getInputOrFail("workspace")
	message := getInputOrFail("message")
	directory := getInputOrFail("directory")

	speculative, err := strconv.ParseBool(getInputOrFail("speculative"))
	if err != nil {
		exitError(err)
	}

	return input{
		token:        token,
		organization: organization,
		workspace:    workspace,
		message:      message,
		directory:    directory,
		speculative:  speculative,
	}
}

func getInputOrFail(identifier string) string {
	val := gha.GetInput(identifier)
	if val == "" {
		exitErrorf("missing input '%v'", identifier)
	}
	return val
}
