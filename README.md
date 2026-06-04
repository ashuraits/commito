# commito

A local git diff viewer that opens in your browser. Stage, unstage, revert files, and browse diffs — without leaving the terminal workflow.

## Usage

```bash
commito             # opens browser at http://localhost:7171
commito -p 8080     # custom port
commito -d /path    # specify repo directory
commito --no-open   # don't open browser automatically
```

## Install

```bash
go install github.com/alexshuraits/commito@latest
```

Or build from source:

```bash
make build
```

## Stack

- **Backend**: Go + embedded Svelte frontend served over HTTP
- **Frontend**: Svelte + TypeScript (built into the binary via `go:embed`)
