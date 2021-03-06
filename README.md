# find-affected-packages

Prints a list of affected Go packages within a given Git commit range, optionally limiting the results to only packages matching a Go package pattern.

Changes to module dependencies (based on changes to the `go.sum` file) are also included in this calculation, so if you change a dependency, any packages using that dependency (recursively) will be considered affected. This takes a pessimistic approach: If you update a module, this assumes all packages inside that module may have changed.

This can be useful to decide which executables to build or which tests to run in a large repository. For example, on a branch, you could run this to test only packages which are affected by changes commited on the current branch since it diverged from master:

```bash
export GIT_RANGE=$(git merge-base origin/master HEAD)..HEAD
go test $(find-affected-packages $GIT_RANGE)
```

## Requirements

This command must be run from the root of your repository, and the root of the repository must also be where a Go module is defined (`go.mod` and `go.sum` must exist there).

## Usage

```find-affected-packages <start commit>..<end commit> [packages...]```

Show all affected packages between two commits:

```find-affected-packages abc123..def456```

Show all affected packages matching `./my-package/...`:

```find-affected-packages abc123..def456 ./my-package/...```