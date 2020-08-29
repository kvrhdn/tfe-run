package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	if gha.InGitHubActions() {
		err = gha.PopulateFromInputs(&input)
	} else {
		err = unmarshalJSON("input.json", &input)
	}
	if err != nil {
		fmt.Printf("Error: could not read input: %v", err)
		os.Exit(1)
	}

	ctx := context.Background()

	options := tferun.RunOptions{
		Token:             input.Token,
		Organization:      input.Organization,
		Workspace:         input.Workspace,
		Message:           notEmptyOrNil(input.Message),
		Directory:         notEmptyOrNil(input.Directory),
		Speculative:       input.Speculative,
		WaitForCompletion: input.WaitForCompletion,
		TfVars:            notEmptyOrNil(input.TfVars),
	}
	output, err := tferun.Run(ctx, options)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	writeOutput(&output)
}

func unmarshalJSON(filename string, v interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read '%v': %w", filename, err)
	}

	err = json.Unmarshal(bytes, &v)
	if err != nil {
		return fmt.Errorf("could not unmarshal JSON '%v': %w", filename, err)
	}
	return nil
}

func writeOutput(output *tferun.RunOutput) {
	if gha.InGitHubActions() {
		gha.WriteOutput("run-url", output.RunURL)
		if output.HasChanges != nil {
			gha.WriteOutput("has-changes", strconv.FormatBool(*output.HasChanges))
		}

		if output.TfOutputs != nil {
			for k, v := range *output.TfOutputs {
				gha.WriteOutput(fmt.Sprintf("tf-%v", k), v)
			}
		}
	} else {
		fmt.Printf("Output:\n")
		fmt.Printf(" - run-url:     %s\n", output.RunURL)
		fmt.Printf(" - has-changes: %v\n", output.HasChanges)

		fmt.Printf(" - tf outputs:\n")
		if output.TfOutputs != nil {
			for k, v := range *output.TfOutputs {
				fmt.Printf("   - %s: %s\n", k, v)
			}
		}
	}
}

func notEmptyOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
