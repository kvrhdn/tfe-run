package main

import (
	"context"
	"fmt"
	"os"

	tfe "github.com/hashicorp/go-tfe"
	gha "github.com/sethvargo/go-githubactions"
)

type input struct {
	token        string
	organization string
	workspace    string
	message      string
	directory    string
	speculative  bool
}

func main() {
	in := readInput()

	config := &tfe.Config{
		Token: in.token,
	}
	client, err := tfe.NewClient(config)
	if err != nil {
		exitError(err)
	}

	ctx := context.Background()

	w, err := client.Workspaces.Read(ctx, in.organization, in.workspace)
	if err != nil {
		gha.Fatalf("%v", err)
	}

	cvOptions := tfe.ConfigurationVersionCreateOptions{
		// Don't automatically queue the runs, we create the run manually to set the message
		AutoQueueRuns: pb(false),
		Speculative:   &in.speculative,
	}
	cv, err := client.ConfigurationVersions.Create(ctx, w.ID, cvOptions)
	if err != nil {
		gha.Fatalf("%v", err)
	}

	err = client.ConfigurationVersions.Upload(ctx, cv.UploadURL, in.directory)
	if err != nil {
		gha.Fatalf("%v", err)
	}

	rOptions := tfe.RunCreateOptions{
		Workspace:            w,
		ConfigurationVersion: cv,
		Message:              &in.message,
	}
	r, err := client.Runs.Create(ctx, rOptions)
	if err != nil {
		gha.Fatalf("%v", err)
	}

	runURL := fmt.Sprintf(
		"https://app.terraform.io/app/%v/workspaces/%v/runs/%v",
		in.organization, in.workspace, r.ID,
	)

	fmt.Printf("Run %v has been queued\n", r.ID)
	fmt.Printf("View the run online: %v\n", runURL)

	gha.SetOutput("run-url", runURL)

	// If auto apply isn't enabled a run can hang for a long time, even if this
	// wouldn't change anything the previous could still be hanging.
	// We don't want to wait for this.
	if w.AutoApply == false {
		fmt.Print("Auto apply is not enabled, won't wait for completion\n")
		return
	}

	var prevStatus tfe.RunStatus
	for {
		r, err = client.Runs.Read(ctx, r.ID)
		if err != nil {
			gha.Fatalf("%v", err)
		}

		if prevStatus != r.Status {
			fmt.Printf("Run status: %v\n", r.Status)
			prevStatus = r.Status
		}

		switch r.Status {

		case tfe.RunPlannedAndFinished:
			exit("Run has been planned, nothing to do")
		case tfe.RunApplied:
			exit("Run has been applied!")

		case tfe.RunCanceled:
			exitErrorf("Run %v has been canceled", r.ID)
		case tfe.RunDiscarded:
			exitErrorf("Run %v has been discarded", r.ID)
		case tfe.RunErrored:
			exitErrorf("Run %v has errored", r.ID)
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
	gha.Fatalf("%v", v)
}

func exitErrorf(msg string, args ...interface{}) {
	gha.Fatalf(msg, args)
}
