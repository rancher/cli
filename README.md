Rancher CLI
===========

The Rancher Command Line Interface (CLI)is a unified tool to manage your Rancher server. With this tool, you can control your services, containers and hosts within a Rancher environment and automate them through scripts. 

## Running

You can check the [releases page](https://github.com/rancher/cli/releases) for direct downloads of the binary or [build your own](#building). 

## Setting up Rancher CLI with Rancher Server 

To enable the CLI to connect to Rancher server, you can configure the environment variables needed. The environment variables that are required are `RANCHER_URL`, `RANCHER_ACCESS_KEY` and `RANCHER_SECRET_KEY`. 

The access key and secret key should be an [account API key](http://docs.rancher.com/rancher/latest/en/api/api-keys/#account-api-keys). In your Rancher setup, you can create an account API key under the **API** tab and expand the **Advanced Options**. 

You can run `rancher config` to set these environment variables for the CLI. 

```
$ rancher --url http://<RANCHER_SERVER_URL> config
URL [http://<RANCHER_SERVER_URL>]: 
Access Key [http://<RANCHER_SERVER_URL>]: <ACCESS_KEY>
Secret Key [http://<RANCHER_SERVER_URL>]: <SECRET_KEY>
INFO[0102] Saving config to /Users/<username>/.rancher/cli.json 
```

> Note: The `<RANCHER_SERVER_URL>` includes whatever port was exposed when installing Rancher server. If you had followed the installation instructions, your URL would be `http://<server_ip>:8080/`.

## Building

The binaries will be located in `/build/bin`.

### Linux binary

Run `make`.

### Mac binary

Run `CROSS=1 make build`


## License
Copyright (c) 2014-2016 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
