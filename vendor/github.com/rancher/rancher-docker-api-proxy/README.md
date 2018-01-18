# Rancher Docker API Proxy

This is very simple library to access the Docker socket through the Rancher API.
This allows one to communicate with Docker without exposing the Docker socket through
TLS or any public port.


## Usage

```go
package main

import (
    "os"

    "github.com/Sirupsen/logrus"
    rancher "github.com/rancher/go-rancher/v2"
)

func main() {
    // Simple example of using this library. Run this as follows
    //
    //     go run main/main.go myhost unix:///tmp/myhost.sock
    //
    // Then run `docker -H unix:///tmp/myhost.sock ps`
    if err := run(); err != nil {
        logrus.Fatal(err)
    }
}

func run() error {
    logrus.SetLevel(logrus.DebugLevel)

    client, err := rancher.NewRancherClient(&rancher.ClientOpts{
        Url:       os.Getenv("CATTLE_URL"),
        AccessKey: os.Getenv("CATTLE_ACCESS_KEY"),
        SecretKey: os.Getenv("CATTLE_SECRET_KEY"),
    })
    if err != nil {
        return err
    }

    proxy := dockerapiproxy.NewProxy(client, .Args(1), os.Args(2))
    return proxy.ListenAndServe()
}
```

Construct a docker API client routing through Rancher.

```go
func NewAPIClient(...) (client.APIClient, error) {
    httpClient, err := newHTTPClient(host, tlsOptions.TLSOptions)
    if err != nil {
        return nil, err
    }

    return client.NewClient(host, verStr, httpClient, customHeaders)
}

func newHTTPClient(host string, tlsOptions *tlsconfig.Options) (*http.Client, error) {
    client, err := rancher.NewRancherClient(&rancher.ClientOpts{
        Url: "http://localhost:8080",
    })
    if err != nil {
        return nil, err
    }


    // second arguement can be host id, name, or hostname
    dialer, err := dockerapiproxy.NewDialer(client, "1h123")
    if err != nil {
        return nil, err
    }

    return &http.Client{
        Transport: &http.Transport{
            Dial: dialer,
        },
    }, nil
}

```



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
