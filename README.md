[![Build Status](https://github.com/azazeal/flycast/actions/workflows/build.yml/badge.svg)](https://github.com/azazeal/flycast/actions/workflows/build.yml)
[![Coverage Report](https://coveralls.io/repos/github/azazeal/flycast/badge.svg?branch=master)](https://coveralls.io/github/azazeal/flycast?branch=master)

# flycast

`flycast` implements persevering UDP broadcasting for apps running on [Fly](http://fly.io).

It does so by accepting UDP packets on two configurable port numbers and by
_re-sending_ those packets to global or regional (depending on the intercepting
port) instances of a configurable target app via the organization's internal
network.

`flycast` discovers the instances it should broadcast to automatically, via 
querying Fly's internal DNS, and very frequently (currently every second) and 
comes with an embedded HTTP server that exports a complete health check.

An example deployment configuration can be found in the 
[`fly.example.toml`](https://github.com/azazeal/flycast/blob/master/fly.example.toml)
file of this repo.

## Disclaimer

- This here program works not, maybe possibly, yet.
- 16384 bytes, or 16 KiB, is the maximum UDP packet size `flycast` can handle.

## Configuration

`flycast` is configured via the following environment variables:

| Variable       | Description                                                                                      | Default value   |
| -------------- | ------------------------------------------------------------------------------------------------ | --------------- |
| `$APP`         | Fly app to broadcast to.                                                                         | `$FLY_APP_NAME` |
| `$PORT_GLOBAL` | Packets arriving on this port will be broadcasted to all instances of `$APP`.                    | `65535`         |
| `$PORT_LOCAL`  | Packets arriving on this port will be broadcasted to instances of `$APP` on the same region.     | `65534`         |
| `$PORT_RELAY`  | `flycast` will broadcast packets to this port.                                                   | `65533`         |
| `$PORT_HTTP`   | The embedded web browser will run on this port with the health check accessible under `/health`. | `8080`          |
| `$LOG_LEVEL`   | Controls the verbosity of the logger. Valid values are `debug`, `info`, `warn`, `error`.         | `info`          |
| `$LOG_FORMAT`  | When set to `json` instructs the logger to output JSON objects instead of raw text.              | N/A             |
