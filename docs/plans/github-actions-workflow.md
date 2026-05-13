# Plan: GitHub Actions Test Workflow

## Context

The repository has no CI/CD configuration. We need a GitHub Actions workflow to run tests automatically on push to `main` and on pull requests, ensuring the library works across supported Go versions. Modeled after [kyma-project/rt-bootstrapper](https://github.com/kyma-project/rt-bootstrapper/actions) workflow (without e2e tests).

## Implementation

Create `.github/workflows/test.yml`:

```yaml
name: Tests

on:
  push:
    branches:
      - main
    paths-ignore:
      - LICENSE
      - .gitignore
      - "**.md"
  pull_request:
    types: [opened, synchronize, reopened]
    paths-ignore:
      - LICENSE
      - .gitignore
      - "**.md"

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Clone the code
        uses: actions/checkout@v5

      - name: Setup Go
        uses: actions/setup-go@v6
        with:
          go-version-file: 'go.mod'

      - name: Run tests
        run: |
          go mod tidy
          set -o pipefail
          go test -v -race ./pkg/... | tee bt-unit-tests.txt

      - name: Upload artifact
        uses: actions/upload-artifact@v5
        with:
          name: bt-unit-tests
          path: bt-unit-tests.txt

      - name: Generate test summary
        run: |
          {
            echo "# bt unit tests output:"
            printf '\n```\n'
            cat bt-unit-tests.txt
            printf '\n```\n'
          } >> $GITHUB_STEP_SUMMARY
```

## Key decisions

- **`paths-ignore`** — skips CI for docs-only and license changes
- **`go-version-file: 'go.mod'`** — always uses the version declared in the module, no manual matrix to maintain
- **`-race` flag** — catches concurrency bugs (relevant for BT use in game/robotics loops)
- **Test output artifact** — preserves full test output for later inspection
- **`$GITHUB_STEP_SUMMARY`** — renders test output directly in the Actions run summary tab
- **No e2e tests** — not needed for a pure library with no external dependencies

## Files to create

- `.github/workflows/test.yml`

## Verification

After creating the file, push a commit and check the Actions tab on GitHub, or validate the workflow syntax locally with `actionlint` if available.
