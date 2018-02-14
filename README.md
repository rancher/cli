Rancher CLI
===========

The Rancher Command Line Interface (CLI) is a unified tool to interact with your Rancher server. 

## Installing

You can check the [releases page](https://github.com/rancher/cli/releases) for direct downloads of the binary or [build your own](#building). 

## Setting up Rancher CLI with Rancher Server 

The CLI needs to know your server address and the credentials required to authenticate with it. 
Rancher CLI will pull this information from a `cli.json` that is created the first time you run 
`rancher login`. By default this file is located at `~/.rancher/cli.json`. 

```
$ rancher login https://<RANCHER_SERVER_URL> -t my-secret-token --name CoolServer1
```

> Note: The `<RANCHER_SERVER_URL>` includes whatever port was exposed when installing Rancher server.

If you want to use Rancher CLI on a server that uses a self signed cert you will need to download the cert from `<RANCHER_SERVER_URL>/v3/settings` and pass that into `rancher login` using `--cacert` 

## Building from source

The binaries will be located in `/bin`.

### Linux binary

Run `make`.

### Mac binary

Run `CROSS=1 make build`

## Docker image

Run `docker run --rm -it rancher/cli [ARGS]`  You can pass in credentials by bind mounting in a config file.

To build `rancher/cli` just run `make`.  To use a custom Docker repository do `REPO=custom make` and it will producte `custom/cli` image.

## Contact

For bugs, questions, comments, corrections, suggestions, etc., open an issue in
[rancher/rancher](//github.com/rancher/rancher/issues) with a title starting with `[cli] `.

Or just [click here](//github.com/rancher/rancher/issues/new?title=%5Bcli%5D%20) to create a new issue.

## License
Copyright (c) 2014-2018 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
