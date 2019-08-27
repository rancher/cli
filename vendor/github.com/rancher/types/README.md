Rancher API Types
========

API Types for Rancher 2.0

## Building

`make`

## Running the code generator

Run `go generate` in the root of the project

### Running code generation after go modules

The code generation code still depends on the GOPATH. 
To generate types you need to update vendor.

1. Update go.mod for what you need 
2. Run `GO111MODULE=on go mod vendor`
3. Export your gopath (export GOPATH=types_go_path_dir)
4. Run `go generate`
5. Unset your `GOPATH`

## License
Copyright (c) 2014-2017 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
