# `tfe-run` Action

![CI](https://github.com/kvrhdn/tfe-run/workflows/CI/badge.svg)
![Integration](https://github.com/kvrhdn/tfe-run/workflows/Integration/badge.svg)

This GitHub Action creates a new run on Terraform Cloud. Integrate Terraform Cloud into your GitHub Actions workflow.

## How to use it

```yaml
- uses: kvrhdn/tfe-run@v1
  with:
    token: ${{ secrets.TFE_TOKEN }}
    workspace: tfe-run
    message: |
      Run triggered using tfe-run (commit: ${{ github.SHA }})
  id: tfe-run

... next steps can access the run URL with ${{ steps.tfe-run.outputs.run-url }}
```

Full option list:

```yaml
- uses: kvrhdn/tfe-run@v1
  with:
    # Token used to communicate with the Terraform Cloud API. Must be a user or
    # team api token.
    token: ${{ secrets.TFE_TOKEN }}

    # Name of the organization on Terraform Cloud. Defaults to the GitHub
    # organization name.
    organization: kvrhdn

    # Name of the workspace on Terraform Cloud.
    workspace: tfe-run

    # Optional message to use as name of the run.
    message: |
      Run triggered using tfe-run (commit: ${{ github.SHA }})

    # The directory that is uploaded to Terraform Enterprise, defaults to the
    # repository root. Respsects .terraformignore
    directory: infrastructure/

    # Whether to run a speculative plan.
    speculative: false

    # The contents of a auto.tfvars file that will be uploaded to Terraform
    # Cloud. This can be used to set Terraform variables.
    tf-vars: |
      run_number = ${{ github.run_number }}

  # Optionally, assign this step an ID so you refer to the outputs from the
  # action with ${{ steps.<id>.outputs.<output variable> }}
  id: tfe-run
```


### Inputs

Name           | Required | Description                                                                                                     | Type   | Default
---------------|----------|-----------------------------------------------------------------------------------------------------------------|--------|--------
`token`        | yes      | Token used to communicating with the Terraform Cloud API. Must be [a user or team api token][tfe-tokens].       | string | 
`organization` | yes      | Name of the organization on Terraform Cloud.                                                                    | string |
`workspace`    | yes      | Name of the workspace on Terraform Cloud.                                                                       | string |
`message`      |          | Optional message to use as name of the run.                                                                     | string | _Queued by GitHub Actions (commit: $GITHUB_SHA)_
`directory`    |          | The directory that is uploaded to Terraform Enterprise, defaults to repository root. Respects .terraformignore. | string | `./`
`speculative`  |          | Whether to run [a speculative plan][tfe-speculative-plan].                                                      | bool   | `false`
`tf-vars`      |          | The contents of a auto.tfvars file that will be uploaded to Terraform Cloud.                                    | string |

[tfe-tokens]: https://www.terraform.io/docs/cloud/users-teams-organizations/api-tokens.html
[tfe-speculative-plan]: https://www.terraform.io/docs/cloud/run/index.html#speculative-plans

### Outputs

Name          | Description                                                                                       | Type
--------------|---------------------------------------------------------------------------------------------------|-----
`run-url`     | URL of the run on Terraform Cloud                                                                 | string
`has-changes` | Whether the run has changes.                                                                      | bool (`'true'` or `'false'`)
`tf-**`       | Outputs from the current Terraform state, prefixed with `tf-`. Only set for non-speculative runs. | string

## License

This Action is distributed under the terms of the MIT license, see [LICENSE](./LICENSE) for details.

## Local development

The easiest way to work on this locally is to run the Go program directly. The program will check whether it is running within the GitHub Actions environment and if not, read its inputs from a file `input.json`.

Create a file `input.json` which contains the inputs that would otherwise be provided by GitHub Actions.

```json
# input.json
{
    "token": "...",
    "organization": "kvrhdn",
    "workspace": "tfe-run-integration",
    "speculative": false,
    "message": "Queued locally using tfe-run",
    "directory": ".",
    "tfVars": "run_number = 0"
}
```

Next, run the program locally:

```
go run .
```
