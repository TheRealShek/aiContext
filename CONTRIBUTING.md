# Contributing to aiContext

## Requirements

- Go 1.25 or newer
- A POSIX shell for validating `install.sh`

## Development workflow

Build the CLI:

```sh
go build ./...
```

Run all local checks:

```sh
gofmt -w main.go main_test.go
go vet ./...
go test ./...
sh -n install.sh
```

Test the CLI without modifying your normal templates or a real project:

```sh
tmp_templates="$(mktemp -d)"
tmp_project="$(mktemp -d)"
go run . setup --template-dir "$tmp_templates"
go run . init --target "$tmp_project" --template-dir "$tmp_templates"
```

Please add or update tests for behavior changes. Keep the CLI dependency-free unless a dependency provides a clear maintenance or correctness benefit.

## Pull requests

Before opening a pull request:

1. Run the formatting, vet, test, and installer syntax checks above.
2. Confirm `git diff --check` reports no whitespace errors.
3. Update the README when commands, generated files, or installation behavior change.

CI runs these checks for pushes and pull requests.

## Releases

Releases are tag-driven. After the desired commit is on `main`, create and push a semantic version tag:

```sh
git tag v1.2.0
git push origin v1.2.0
```

The release workflow validates the code again, then GoReleaser builds platform archives and publishes checksums. Do not reuse or move an existing release tag.
