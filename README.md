Rancher CLI
===========
>[!NOTE]
>We are currently not accepting feature contributions from the community


The Rancher Command Line Interface (CLI) is a unified tool for interacting with your Rancher Server.

For usage information see: https://rancher.com/docs/rancher/v2.x/en/cli/

> **Note:** This is for version 2.x.x of the cli, for info on 1.6.x see [here](https://github.com/rancher/cli/tree/v1.6)

## Installing

Check the [releases page](https://github.com/rancher/cli/releases) for direct downloads of the binary. After you download it, you can add it to your `$PATH` or [build your own from source](#building-from-source).

## Setting up Rancher CLI with a Rancher Server

The CLI requires your Rancher Server address, along with [credentials for authentication](https://rancher.com/docs/rancher/v2.x/en/user-settings/api-keys/). Rancher CLI pulls this information from a JSON file, `cli2.json`, which is created the first time you run `rancher login`. By default, the path of this file is `~/.rancher/cli2.json`.

```
$ rancher login https://<RANCHER_SERVER_URL> -t my-secret-token
```

> **Note:** When entering your `<RANCHER_SERVER_URL>`, include the port that was exposed while you installed Rancher Server.

## Usage

Run `rancher --help` for a list of available commands.

## Building from Source

The binaries will be located in `/bin`.

### Linux Binary

Run `make build`.

### Mac Binary

Run `CROSS=1 make build`.

## Docker Image

Run `docker run --rm -it -v <PATH_TO_CONFIG>:/home/cli/.rancher/cli2.json rancher/cli2 [ARGS]`.
Pass credentials by replacing `<PATH_TO_CONFIG>` with your config file for the server.

To build `rancher/cli`, run `make`.  To use a custom Docker repository, do `REPO=custom make`, which produces a `custom/cli` image.

## Contact

For bugs, questions, comments, corrections, suggestions, etc., open an issue in
[rancher/rancher](//github.com/rancher/rancher/issues) with a title prefix of `[cli] `.

Or just [click here](//github.com/rancher/rancher/issues/new?title=%5Bcli%5D%20) to create a new issue.

## License
Copyright (c) 2014-2019 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
