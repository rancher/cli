
rancher-catalog-service
=======================
REST Service enables a user to view a catalog of pre-cooked templates stored on a github repo. Also the user can  launch the templates onto a specific Rancher environment.

Design
==========
* rancher-catalog-service gets deployed as a Rancher service containerized app. 

* rancher-catalog-service will clone a public github repo and provide API to list and navigate through the templates and subversions from the repo

* The service will periodically sync  changes from the repo

* The UI integrated with the service will enable the user to view the templates in a catalog format and also launch a template to a specified rancher deployment.

Building
========

This project uses [dapper](https://github.com/ibuildthecloud/dapper).  Install dapper first

    go get github.com/ibuildthecloud/dapper

```sh
# Compile
dapper build

# Run tests
dapper test

# Run everything
dapper all
```

Contact
========
For bugs, questions, comments, corrections, suggestions, etc., open an issue in
 [rancher/rancher](//github.com/rancher/rancher/issues).

Or just [click here](//github.com/rancher/rancher/issues/new?title=%5Brancher-dns%5D%20) to create a new issue.

License
=======
Copyright (c) 2015 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
