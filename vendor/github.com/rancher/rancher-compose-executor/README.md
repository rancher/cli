# Rancher Compose [![Build Status](http://ci.rancher.io/api/badge/github.com/rancher/rancher-compose/status.svg?branch=master)](http://ci.rancher.io/github.com/rancher/rancher-compose)

Docker compose compatible client that deploys to [Rancher](https://github.com/rancher/rancher).

## Binaries

Binaries are available for Linux, OS X, and Windows. Refer to the latest [release](https://github.com/rancher/rancher-compose/releases).

## Building
Run `make build` to create `./bin/rancher-compose`

## Usage:

```
Usage: rancher-compose [OPTIONS] COMMAND [arg...]

Docker-compose to Rancher

Version: v0.8.4

Author:
Rancher Labs, Inc.

Options:
--verbose, --debug				
--file, -f [--file option --file option]	Specify one or more alternate compose files (default: docker-compose.yml) [$COMPOSE_FILE]
--project-name, -p 				Specify an alternate project name (default: directory name)
--url 					Specify the Rancher API endpoint URL [$RANCHER_URL]
--access-key 					Specify Rancher API access key [$RANCHER_ACCESS_KEY]
--secret-key 					Specify Rancher API secret key [$RANCHER_SECRET_KEY]
--rancher-file, -r 				Specify an alternate Rancher compose file (default: rancher-compose.yml)
--env-file, -e 				Specify a file from which to read environment variables
--help, -h					show help
--version, -v					print the version

Commands:
create	Create all services but do not start
up		Bring all services up
start		Start services
logs		Get service logs
restart	Restart services
stop, down	Stop services
scale		Scale services
rm		Delete services
pull		Pulls images for services
upgrade	Perform rolling upgrade between services
help, h	Shows a list of commands or help for one command
```

# Compose compatibility

`rancher-compose` strives to be completely compatible with Docker Compose.  Since `rancher-compose` is largely focused
on running production workloads some behaviors between Docker Compose and Rancher Compose are different.

## Deleting Services/Container

`rancher-compose` will not delete things by default.  This means that if you do two `up` commands in a row, the second `up` will
do nothing.  This is because the first up will create everything and leave it running.  Even if you do not pass `-d` to `up`,
`rancher-compose` will not delete your services.  To delete a service you must use `rm`.

## Builds

Docker builds are supported in two ways.  First is to set `build:` to a git or HTTP URL that is compatible with the remote parameter in https://docs.docker.com/reference/api/docker_remote_api_v1.18/#build-image-from-a-dockerfile.  The second approach is to set `build:` to a local directory and the build context will be uploaded to S3 and then built on demand on each node.

For S3 based builds to work you must [setup AWS credentials](https://github.com/aws/aws-sdk-go/#configuring-credentials).


## Contact
For bugs, questions, comments, corrections, suggestions, etc., open an issue in
 [rancher/rancher](//github.com/rancher/rancher/issues) with a title starting with `[rancher-compose] `.

Or just [click here](//github.com/rancher/rancher/issues/new?title=%5Brancher-compose%5D%20) to create a new issue.

# License
Copyright (c) 2014-2015 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

