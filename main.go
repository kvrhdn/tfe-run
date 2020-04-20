package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	tfe "github.com/hashicorp/go-tfe"
	gha "github.com/sethvargo/go-githubactions"
)

type input struct {
	Token        string `json:"token"`
	Organization string `json:"organization"`
	Workspace    string `json:"workspace"`
	Message      string `json:"message"`
	Directory    string `json:"directory"`
	Speculative  bool   `json:"speculative"`
}

func main() {
	in := readInput()

	config := &tfe.Config{
		Token: in.Token,
	}
	client, err := tfe.NewClient(config)
	if err != nil {
		exitErrorf("Could not create a new TFE client: %v\n", err)
	}

	ctx := context.Background()

	w, err := client.Workspaces.Read(ctx, in.Organization, in.Workspace)
	if err != nil {
		exitErrorf("Could not retrieve workspace '%v/%v': %v\n", in.Organization, in.Workspace, err)
	}

	cvOptions := tfe.ConfigurationVersionCreateOptions{
		// Don't automatically queue the runs, we create the run manually to set the message
		AutoQueueRuns: pb(false),
		Speculative:   &in.Speculative,
	}
	cv, err := client.ConfigurationVersions.Create(ctx, w.ID, cvOptions)
	if err != nil {
		if err.Error() == "resource not found" {
			exitErrorf("Could not create configuration version (not found), this might happen if you are not using a user or team API token\n")
		} else {
			exitErrorf("Could not create a new configuration version: %v\n", err)
		}
	}

	fmt.Print("Uploading directory...\n")

	err = client.ConfigurationVersions.Upload(ctx, cv.UploadURL, in.Directory)
	if err != nil {
		exitErrorf("Could not upload directory '%v': %v\n", in.Directory, err)
	}

	rOptions := tfe.RunCreateOptions{
		Workspace:            w,
		ConfigurationVersion: cv,
		Message:              &in.Message,
	}
	r, err := client.Runs.Create(ctx, rOptions)
	if err != nil {
		exitErrorf("%v", err)
	}

	runURL := fmt.Sprintf(
		"https://app.terraform.io/app/%v/workspaces/%v/runs/%v",
		in.Organization, in.Workspace, r.ID,
	)

	fmt.Printf("Run %v has been queued\n", r.ID)
	fmt.Printf("View the run online: %v\n", runURL)

	gha.SetOutput("run-url", runURL)

	// If auto apply isn't enabled a run could hang for a long time, even if
	// the run itself wouldn't change anything the previous run could still be
	// blocked waiting for confirmation.
	// Speculative runs can always continue it seems.
	if !in.Speculative && !w.AutoApply {
		fmt.Print("Auto apply isn't enabled, won't wait for completion.\n")
		return
	}

	var prevStatus tfe.RunStatus
	for {
		r, err = client.Runs.Read(ctx, r.ID)
		if err != nil {
			exitErrorf("Could not read run '%v': %v\n", r.ID, err)
		}

		if prevStatus != r.Status {
			fmt.Printf("Run status: %v\n", r.Status)
			prevStatus = r.Status
		}

		switch r.Status {

		case tfe.RunPlannedAndFinished:
			gha.SetOutput("has-changes", strconv.FormatBool(r.HasChanges))
			exit("Run has been planned, nothing to do\n")
		case tfe.RunApplied:
			gha.SetOutput("has-changes", strconv.FormatBool(r.HasChanges))
			exit("Run has been applied!\n")

		case tfe.RunCanceled:
			exitErrorf("Run %v has been canceled\n", r.ID)
		case tfe.RunDiscarded:
			exitErrorf("Run %v has been discarded\n", r.ID)
		case tfe.RunErrored:
			exitErrorf("Run %v has errored\n", r.ID)
		}
	}
}

func pb(value bool) *bool {
	local := value
	return &local
}

func exit(v interface{}) {
	fmt.Print(v)
	os.Exit(0)
}

func exitError(v interface{}) {
	fmt.Print(v)
	os.Exit(1)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
	os.Exit(1)
}
