# `tfe-run` Action

![CI](https://github.com/kvrhdn/tfe-run/workflows/CI/badge.svg)
![Integration](https://github.com/kvrhdn/tfe-run/workflows/Integration/badge.svg)

This GitHub Action creates a new run on Terraform Cloud. Integrate Terraform Cloud into your GitHub Actions workflow.

⚠️ This action is still in development, for now use the version from the master branch `kvrhdn/tfe-run@master`. I plan to introduce a `v1` tag eventually.

## How to use it

```yaml
name: Build & deploy
on: [push]

jobs:
  build:
    name: Build & deploy
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v2

      ... other build steps ...

      - uses: kvrhdn/tfe-run@master
        with:
          token: ${{ secrets.TFE_TOKEN }}
          workspace: tfe-run
          message: |
            Run triggered using tfe-run (commit: ${{ github.SHA }})
        id: tfe-run

      ... next steps can access the run URL with ${{ steps.tfe-run.outputs.run-url }}
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

Name          | Description                       | Type
--------------|-----------------------------------|-----
`run-url`     | URL of the run on Terraform Cloud | string
`has-changes` | Whether the run has changes.      | bool (`'true'` or `'false'`)

## How does it work?

This action will interact with [the Terraform Cloud API][tf-cloud-api] to manually create a new run.

First, it will look up the workspace and create a new [_Configuration Version_][tfe-api-configuration-version]. Next the contents of `directory` are uploaded to this configuration version (if `directory` is not specified, the repository will be uploaded). Uploading respects the [`.terraformignore`][terraformignore] file.

Lastly, [a run][tfe-api-run] is created linked to the configuration version, this allows setting a custom message.

If the workspace has [auto apply enabled][tfe-auto-apply], the action will keep track of the scheduled run until it has completed or failed. If auto apply is not enalbled, the action will return immediately to avoid hanging.

[tf-cloud-api]: https://www.terraform.io/docs/cloud/run/api.html
[tfe-api-configuration-version]: https://www.terraform.io/docs/cloud/api/configuration-versions.html
[tfe-api-run]: https://www.terraform.io/docs/cloud/api/run.html
[terraformignore]: https://www.terraform.io/docs/backends/types/remote.html#excluding-files-from-upload-with-terraformignore
[tfe-auto-apply]: https://www.terraform.io/docs/cloud/workspaces/settings.html#auto-apply-and-manual-apply

## License

This Action is distributed under the terms of the MIT license, see [LICENSE](./LICENSE) for details.

## Local development

Easiest way to work on this locally is to run the Go program directly.

First create a file `input.json` which contains the inputs which are otherwise provided by GitHub Actions. Make sure _all inputs_ are present, missing inputs will cause inconsist errors.

```json
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
