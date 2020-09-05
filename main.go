package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	tferun "github.com/kvrhdn/go-tfe-run"
	"github.com/kvrhdn/tfe-run/gha"
)

type input struct {
	Token             string `gha:"token,required"`
	Organization      string `gha:"organization,required"`
	Workspace         string `gha:"workspace,required"`
	Message           string
	Directory         string
	Speculative       bool
	WaitForCompletion bool   `gha:"wait-for-completion"`
	TfVars            string `gha:"tf-vars"`
}

func main() {
	var input input
	var err error

	if !gha.InGitHubActions() {
		exitWithError(errors.New("tfe-run should only be run within GitHub Actions"))
	}

	err = gha.PopulateFromInputs(&input)
	if err != nil {
		exitWithError(fmt.Errorf("could not read inputs: %w", err))
	}

	ctx := context.Background()

	cfg := tferun.ClientConfig{
		Token:        input.Token,
		Organization: input.Organization,
		Workspace:    input.Workspace,
	}
	c, err := tferun.NewClient(ctx, cfg)
	if err != nil {
		exitWithError(err)
	}

	options := tferun.RunOptions{
		Message:           notEmptyOrNil(input.Message),
		Directory:         notEmptyOrNil(input.Directory),
		Speculative:       input.Speculative,
		WaitForCompletion: input.WaitForCompletion,
		TfVars:            notEmptyOrNil(input.TfVars),
	}
	output, err := c.Run(ctx, options)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	gha.WriteOutput("run-url", output.RunURL)
	if output.HasChanges != nil {
		gha.WriteOutput("has-changes", strconv.FormatBool(*output.HasChanges))
	}

	outputs, err := c.GetTerraformOutputs(ctx)
	if err != nil {
		exitWithError(err)
	}

	for k, v := range outputs {
		gha.WriteOutput(fmt.Sprintf("tf-%v", k), v)
	}
}

func notEmptyOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func exitWithError(err error) {
	fmt.Printf("Error: %v", err)
	os.Exit(1)
}
