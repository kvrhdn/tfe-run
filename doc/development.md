# Development

## Local development

The easiest way to work on `tfe-run` locally is to run the Go program directly. `tfe-run` will detect whether it is running within the GitHub Actions environment and if not, read its inputs from a file `input.json`.

Create a file `input.json` which contains the inputs that would otherwise be provided by GitHub Actions.

```json
# input.json
{
    "token": "...",
    "organization": "kvrhdn",
    "workspace": "tfe-run_integration",
    "speculative": false,
    "waitForCompletion": true,
    "message": "Queued locally using tfe-run",
    "directory": "./",
    "tfVars": "run_number = 0"
}
```

Next, run the program locally:

```
go run .
```
