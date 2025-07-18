# HAR Server Replay

A CLI tool to replay HTTP responses from a HAR (HTTP Archive) file.

## Description

This tool creates a local HTTP server that serves responses from a HAR file sequentially. It's useful for mocking APIs and testing client applications.

## Installation

Make sure you have Go installed. You can download it from [https://go.dev/](https://go.dev/).

```bash
go install ./cmd/har_server_replay
```

## Usage
``` bash
har_server_replay --har-file <path_to_har_file> --port <port_number>
```


## Options
`--har-file <path_to_har_file>:` Path to the HAR file. Required.

`--port <port_number>:` Port to listen on. Defaults to 8080.

`--verbose`

## Example
```bash
har_server_replay --har-file testdata/sample.har --port 8080
```

## Contributing
Contributions are welcome! Please submit a pull request.
