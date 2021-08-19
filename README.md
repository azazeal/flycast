[![Build Status](https://github.com/azazeal/flycast/actions/workflows/build.yml/badge.svg)](https://github.com/azazeal/flycast/actions/workflows/build.yml)
[![Coverage Report](https://coveralls.io/repos/github/azazeal/flycast/badge.svg?branch=master)](https://coveralls.io/github/azazeal/flycast?branch=master)

# flycast

`flycast` implements persevering UDP broadcasting for apps running on [Fly](http://fly.io).

## Disclaimer

- This here app works not yet, maybe possibly.
- 32768 bytes, or 32 KiB, is the maximum UDP packet size `flycast` can handle.

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