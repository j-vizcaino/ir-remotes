# Broadlink IR blaster tool

[![Build badge]][Build] [![GoReport badge]][GoReport]

[Build badge]: https://travis-ci.org/j-vizcaino/ir-remotes.svg
[Build]: https://travis-ci.org/j-vizcaino/ir-remotes
[GoReport badge]: https://goreportcard.com/badge/github.com/j-vizcaino/ir-remotes
[GoReport]: https://goreportcard.com/report/github.com/j-vizcaino/ir-remotes


`ir-remotes` provides a command line tool for managing Broadlink devices, recording infra-red codes, then, providing a way to replay those, using a REST API endpoint.

## Getting started

Disclaimer: the following code has been tested on Linux, using a Broadlink RM Mini IR blaster. While it may support other Broadlink devices, this has not been tested. Contributions are welcome.

### Installation

To build from source, Go `>= 1.11` is required, since the repository uses Go modules.

```bash
# Clone the repository anywhere you want
$ git clone https://github.com/j-vizcaino/ir-remotes.git
$ cd ir-remotes
$ go build
```

### Discovering devices

The first step is to discover and save the Broadlink devices living in your local network.

```bash
$ ir-remotes devices discover
```

By default, the devices information get stored in `devices.json` but this can be configured using the `--devices-file` option.

When `devices.json` exist, the command preserves its content. It is safe to run the `discover` command many times without loosing previously discovered devices.

### Capturing IR codes

Once device list is ready, it's time to capture some IR codes. Using this mode, the Broadlink device will wait for IR code and record it.

```bash
# Example: record the power, vol_up, vol_down and mute buttons from the TV remote
$ ir-remotes capture -n tv power vol_up vol_down mute
```

The `capture` command

* stores the raw IR codes in the `remotes.json` file (configurable with `--remotes-file` option)
* captures the IR codes sequentially, asking the user to press the IR remote button when ready
* skips already captured IR codes that may exist in the remotes file

### REST endpoint

With device list and a couple of IR codes saved to disk, the REST service can be started.

```bash
$ ir-remotes server
```

The following endpoints are provided by the service:

* `GET /api/devices`: list of Broadlink devices available and listed in the `devices.json`
* `GET /api/devices/:name`: get information for the device with `name`
* `GET /api/remotes`: list of remote names, loaded from `remotes.json`
* `GET /api/remotes/:name`: get the list of IR codes for the remote with `name`
* `POST /api/remotes/:name/:code`: send the IR code named `code`

### All-in-one REST server and web frontend

The `server` command allows for serving static content from disk.
By default, accessing `localhost:8080` redirects to `localhost:8080/ui/` where the application serves the UI asset content, located in `assets/ui`.

In order to improve installation and distribution, a binary can be compiled, embedding both the UI files as well as the `remotes.json` and `devices.json` located in `assets/config` directory.

To build this version of the binary

```bash
$ GO_BUILD_TAGS=embedded make build
```

The resulting `ir-remotes` binary can now be copied anywhere without any additional file.

