// Package io contains all input and output related functionality for tfe-run.
package io

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/kvrhdn/tfe-run/gha"
)

// Input data needed to run tfe-run
type Input struct {
	Token        string `json:"token"`
	Organization string `json:"organization"`
	Workspace    string `json:"workspace"`
	Message      string `json:"message"`
	Directory    string `json:"directory"`
	Speculative  bool   `json:"speculative"`
}

// ReadInput reads and parses input data.
func ReadInput() (Input, error) {
	if gha.InGitHubActions() {
		return readInputGitHubActions()
	}
	return readInputLocal()
}

func readInputGitHubActions() (input Input, err error) {
	input.Token, err = gha.ReadInput("token")
	if err != nil {
		err = fmt.Errorf("could not read input paramater 'token': %w", err)
		return
	}
	gha.AddMask(input.Token)

	input.Organization, err = gha.ReadInput("organization")
	if err != nil {
		err = fmt.Errorf("could not read input paramater 'organization': %w", err)
		return
	}

	input.Workspace, err = gha.ReadInput("workspace")
	if err != nil {
		err = fmt.Errorf("could not read input paramater 'workspace': %w", err)
		return
	}

	input.Message, err = gha.ReadInput("message")
	if err != nil {
		err = fmt.Errorf("could not read input paramater 'message': %w", err)
		return
	}

	input.Directory, err = gha.ReadInput("directory")
	if err != nil {
		err = fmt.Errorf("could not read input paramater 'directory': %w", err)
		return
	}

	speculative, err := gha.ReadInput("speculative")
	if err != nil {
		err = fmt.Errorf("could not read input paramater 'speculative': %w", err)
		return
	}
	input.Speculative, err = strconv.ParseBool(speculative)
	if err != nil {
		err = fmt.Errorf("could not parse input paramater 'speculative' as bool: %w", err)
	}

	return
}

func readInputLocal() (Input, error) {
	bytes, err := ioutil.ReadFile("input.json")
	if err != nil {
		err = fmt.Errorf("could not read 'input.json': %w", err)
		return Input{}, err
	}

	var input Input
	err = json.Unmarshal(bytes, &input)
	if err != nil {
		err = fmt.Errorf("could not parse 'input.json': %w", err)
		return Input{}, err
	}
	return input, nil
}
