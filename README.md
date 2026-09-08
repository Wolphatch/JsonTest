# json-test

`json-test` is a small, dependency-light CLI that sequentially compares JSON returned by baseline and candidate curl requests. It compares objects recursively and, by default, arrays by index, reporting values, types, missing fields, lengths, and HTTP status codes.

## Build

Go 1.22 or newer and `curl` on `PATH` are required at runtime.

```sh
go build -o json-test ./cmd/json-test
```

Cross-compile from any Go host (CGO is not required):

```sh
GOOS=linux   GOARCH=amd64 go build -o json-test-linux-amd64 ./cmd/json-test
GOOS=darwin  GOARCH=arm64 go build -o json-test-darwin-arm64 ./cmd/json-test
GOOS=windows GOARCH=amd64 go build -o json-test-windows-amd64.exe ./cmd/json-test
```

## Commands

```text
json-test run <manifest.yaml>
json-test validate <manifest.yaml>
json-test version
```

`run` exits 0 when every case matches, 1 on a difference/request failure, and 2 for invalid configuration or usage. `validate` parses paths and curl commands without making requests.

## Manifest

See [`examples/basic.yaml`](examples/basic.yaml). `baseline` and `candidate` accept either a curl string or a mapping containing `curl`. Settings at the manifest level are defaults; case-level `timeout`, `compareStatus`, and `compareListOrder` override them. The default timeout is 30 seconds, default report is `text`, and status and list-order comparison both default to true. Set `compareListOrder: false` to treat arrays as unordered (including duplicate values). Set `report: json` for deterministic, machine-readable output.

Includes are a whitelist. Excludes are a blacklist and always win. Dot paths select nested fields, and `[*]` selects every array element (for example, `items[*].price`). Numeric selectors such as `items[0].price` are also accepted.

Only `curl` is launched, directly through Go's process API—never through a shell. Single/double quoting and backslash escaping are supported. Pipes, redirects, chaining, backticks, and `$()` command substitution are rejected. `${ENV_VAR}` placeholders are expanded immediately before execution; missing variables fail the case. Reports never contain curl commands or header values, so credentials are not printed or persisted by json-test. Avoid curl verbose/trace flags because curl itself may emit headers to stderr.

## Development

```sh
go test ./...
go vet ./...
```

The integration test starts a local HTTP server and invokes the real `curl` executable.
