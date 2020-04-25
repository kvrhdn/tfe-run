package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/kvrhdn/tfe-run/io"
)

func main() {
	input, err := io.ReadInput()
	if err != nil {
		fmt.Printf("Error: could not read input: %v", err)
		os.Exit(1)
	}

	output, err := run(input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	io.WriteOutput(&output)
}

func run(input io.Input) (output io.Output, err error) {
	config := &tfe.Config{
		Token: input.Token,
	}
	client, err := tfe.NewClient(config)
	if err != nil {
		err = fmt.Errorf("could not create a new TFE client: %w", err)
		return
	}

	ctx := context.Background()

	w, err := client.Workspaces.Read(ctx, input.Organization, input.Workspace)
	if err != nil {
		err = fmt.Errorf("could not retrieve workspace '%v/%v': %w", input.Organization, input.Workspace, err)
		return
	}

	cvOptions := tfe.ConfigurationVersionCreateOptions{
		// Don't automatically queue the runs, we create the run manually to set the message
		AutoQueueRuns: tfe.Bool(false),
		Speculative:   &input.Speculative,
	}
	cv, err := client.ConfigurationVersions.Create(ctx, w.ID, cvOptions)
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			err = fmt.Errorf("could not create configuration version (404 not found), this might happen if you are not using a user or team API token")
		} else {
			err = fmt.Errorf("could not create a new configuration version: %w", err)
		}
		return
	}

	if input.TfVars != "" {
		// Creating a *.auto.tfvars file is the easiest way to temporarily set a variable. The API
		// exposed by Terraform Cloud only allows setting workspace variables. These variables are
		// persistent across runs which might cause undesired side-effects.
		varsFile := filepath.Join(input.Directory, w.WorkingDirectory, "run.auto.tfvars")

		fmt.Printf("Creating variables file %v\n", varsFile)

		err = ioutil.WriteFile(varsFile, []byte(input.TfVars), 0644)
		if err != nil {
			err = fmt.Errorf("could not write run.auto.tfvars: %w", err)
			return
		}

		defer func() {
			err := os.Remove(varsFile)
			if err != nil {
				fmt.Printf("Could not remove run.auto.tfvars, this might cause issues with later steps: %v", err)
			}
		}()
	}

	fmt.Print("Uploading directory...\n")

	err = client.ConfigurationVersions.Upload(ctx, cv.UploadURL, input.Directory)
	if err != nil {
		err = fmt.Errorf("could not upload directory '%v': %w", input.Directory, err)
		return
	}

	fmt.Print("Done uploading.\n")

	rOptions := tfe.RunCreateOptions{
		Workspace:            w,
		ConfigurationVersion: cv,
		Message:              &input.Message,
	}
	r, err := client.Runs.Create(ctx, rOptions)
	if err != nil {
		err = fmt.Errorf("could not create run: %w", err)
		return
	}

	runURL := fmt.Sprintf(
		"https://app.terraform.io/app/%v/workspaces/%v/runs/%v",
		input.Organization, input.Workspace, r.ID,
	)

	fmt.Printf("Run %v has been queued\n", r.ID)
	fmt.Printf("View the run online: %v\n", runURL)

	output.RunURL = runURL

	// If auto apply isn't enabled a run could hang for a long time, even if
	// the run itself wouldn't change anything the previous run could still be
	// blocked waiting for confirmation.
	// Speculative runs can always continue it seems.
	if !input.Speculative && !w.AutoApply {
		fmt.Print("Auto apply isn't enabled, won't wait for completion.\n")
		return
	}

	var prevStatus tfe.RunStatus
	for {
		r, err = client.Runs.Read(ctx, r.ID)
		if err != nil {
			err = fmt.Errorf("could not read run '%v': %v", r.ID, err)
			return
		}

		if prevStatus != r.Status {
			fmt.Printf("Run status: %v\n", prettyPrint(r.Status))
			prevStatus = r.Status
		}

		if isEndStatus(r.Status) {
			break
		}
	}

	output.HasChanges = r.HasChanges

	switch r.Status {

	case tfe.RunPlannedAndFinished:
		fmt.Println("Run has been planned, nothing to do.")
	case tfe.RunApplied:
		fmt.Println("Run has been applied!")

	case tfe.RunCanceled:
		err = fmt.Errorf("run %v has been canceled", r.ID)
	case tfe.RunDiscarded:
		err = fmt.Errorf("run %v has been discarded", r.ID)
	case tfe.RunErrored:
		err = fmt.Errorf("run %v has errored", r.ID)
	}

	if err != nil {
		return
	}

	if !input.Speculative {
		output.TfOutputs, err = retrieveOutputs(ctx, client, w.ID)
	}

	return
}

func isEndStatus(r tfe.RunStatus) bool {
	// All run statuses: https://github.com/hashicorp/go-tfe/blob/v0.7.0/run.go#L46
	switch r {
	case
		tfe.RunPlannedAndFinished,
		tfe.RunApplied,
		tfe.RunCanceled,
		tfe.RunDiscarded,
		tfe.RunErrored:
		return true

	case
		tfe.RunPlanQueued,
		tfe.RunPlanning,
		tfe.RunPlanned,
		tfe.RunPending,
		tfe.RunConfirmed,
		tfe.RunApplyQueued,
		tfe.RunApplying:
		return false

	case
		tfe.RunCostEstimating,
		tfe.RunCostEstimated,
		tfe.RunPolicyChecked,
		tfe.RunPolicyChecking,
		tfe.RunPolicyOverride,
		tfe.RunPolicySoftFailed:
		fmt.Printf("Run is in unexpected / unsupported status %v, finishing process", r)
		return true
	}
	return true
}

func prettyPrint(r tfe.RunStatus) string {
	return strings.ReplaceAll(fmt.Sprintf("%v", r), "_", " ")
}

type minimalTerraformState struct {
	Outputs map[string]terraformOutput `json:"outputs"`
}

type terraformOutput struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func retrieveOutputs(ctx context.Context, client *tfe.Client, workspaceID string) (outputs map[string]string, err error) {
	s, err := client.StateVersions.Current(ctx, workspaceID)
	if err != nil {
		err = fmt.Errorf("could not fetch current state: %w", err)
		return
	}

	bytes, err := client.StateVersions.Download(ctx, s.DownloadURL)
	if err != nil {
		err = fmt.Errorf("could not download state version: %w", err)
		return
	}

	var state minimalTerraformState
	err = json.Unmarshal(bytes, &state)
	if err != nil {
		err = fmt.Errorf("could not parse state version: %w", err)
		return
	}

	outputs = map[string]string{}
	for k, v := range state.Outputs {
		outputs[k] = v.Value
	}

	fmt.Printf("Outputs from current state:\n")
	for k, v := range outputs {
		fmt.Printf(" - %v: %v\n", k, v)
	}

	return
}
