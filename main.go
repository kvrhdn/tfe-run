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
	Type              string
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

	runType := asRunType(input.Type)

	// Speculative is deprecated, but if type is apply (the default) we still respect it
	if runType == tferun.RunTypeApply && input.Speculative {
		runType = tferun.RunTypePlan
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
		Type:              runType,
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

func asRunType(s string) tferun.RunType {
	switch s {
	case "apply":
		return tferun.RunTypeApply
	case "plan":
		return tferun.RunTypePlan
	case "destroy":
		return tferun.RunTypeDestroy
	}
	exitWithError(fmt.Errorf("Type \"%s\" is not supported, must be plan, apply or destroy", s))
	return 0
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
