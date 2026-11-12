# Rancher CLI

The Rancher Command Line Interface (CLI) manages
[Rancher](https://github.com/rancher/rancher) servers from a terminal: log in,
switch between servers, clusters and projects, manage clusters, projects,
namespaces and members, and authenticate `kubectl` against Rancher-managed
clusters.

> [!IMPORTANT]
> **[rancher-cli](https://github.com/rancher/rancher-cli) is the successor to this CLI and will replace it. Start using it today.**
> It is an experimental, AI-native rewrite for both operators and AI coding agents.
> It adds structured JSON/YAML output, a server-side dry run before every change,
> machine-readable errors and exit codes, and an agent skill. It needs Rancher 2.14 or later.
>
> If you use this CLI only as the kubectl credential plugin, rancher-cli is a drop-in replacement.
> Kubeconfigs downloaded from Rancher keep working after you switch binaries.
> For everything else, the [migration guide](https://github.com/rancher/rancher-cli/blob/main/MIGRATING.md)
> maps every command in this CLI to its replacement.
>
> This CLI remains fully supported until an official deprecation notice.

[Install](#install) · [Quickstart](#quickstart) · [Commands](#commands) ·
[kubectl](#kubectl-integration) · [Configuration](#configuration) ·
[Versioning](#versioning) · [rancher-cli](#using-it-alongside-rancher-cli)

User documentation:
[Rancher CLI](https://ranchermanager.docs.rancher.com/reference-guides/cli-with-rancher/rancher-cli)
in the Rancher Manager docs.

## Install

Download a binary from the
[releases page](https://github.com/rancher/cli/releases). Each release has
archives for:

| OS      | Architectures                    | Archive                                                  |
|---------|----------------------------------|----------------------------------------------------------|
| Linux   | `amd64`, `arm64`, `arm`, `s390x` | `rancher-linux-<arch>-<version>.tar.gz` (or `.tar.xz`)   |
| macOS   | `amd64`, `arm64`                 | `rancher-darwin-<arch>-<version>.tar.gz` (or `.tar.xz`)  |
| Windows | `amd64`, `386`                   | `rancher-windows-<arch>-<version>.zip`                   |

`sha256sum.txt` in each release lists the checksums. The same files are
published to `https://releases.rancher.com/cli2/<version>/`.

Extract the archive and put `rancher` on your `PATH`:

```
curl -fsSLO https://github.com/rancher/cli/releases/download/v2.15.3/rancher-linux-amd64-v2.15.3.tar.gz
tar xzf rancher-linux-amd64-v2.15.3.tar.gz
sudo install rancher-v2.15.3/rancher /usr/local/bin/rancher
```

### Container image

The image is `rancher/cli2`, tagged with the release version. Mount a config
directory to keep the login between runs:

```
docker run --rm -it -v "$HOME/.rancher:/home/cli/.rancher" rancher/cli2:v2.15.3 clusters ls
```

The container runs as UID 1000, which must be able to read and write the
mounted directory.

## Quickstart

1. In the Rancher UI, create an API key: **Account & API Keys** →
   **Create API Key**. Copy the bearer token.
2. Log in. The CLI lists your projects and asks you to pick one as the
   current context:

   ```
   rancher login https://rancher.example.com --token token-abc12:xyz...
   ```

   Options: `--cacert <file>` for a server certificate signed by a private CA,
   `--context <cluster-id>:<project-id>` to skip the prompt, `--name <name>` to
   name the server in the local config.
3. Work with the server:

   ```
   rancher clusters ls
   rancher context switch
   rancher clusters kubeconfig <cluster> > ~/.kube/<cluster>.yaml
   rancher kubectl get pods -A
   ```

## Commands

Run `rancher --help` for the full list and `rancher <command> --help` for
subcommands and flags.

| Command                   | Purpose                                                                              |
|---------------------------|--------------------------------------------------------------------------------------|
| `login`                   | Log in to a Rancher server and save the credentials.                                 |
| `server`                  | List, switch and delete saved servers.                                               |
| `context`                 | Show or switch the current cluster and project.                                      |
| `clusters`                | List, create, import and delete clusters; get a kubeconfig; manage cluster members. |
| `projects`                | List, create and delete projects; manage project members.                            |
| `namespaces`              | List, create, delete and move namespaces in the current project.                     |
| `nodes`, `machines`       | List and delete nodes; list machines.                                                |
| `ps`                      | List workloads in a project.                                                         |
| `inspect`                 | Print a resource as JSON or YAML.                                                    |
| `settings`                | List, get and set Rancher settings.                                                  |
| `wait`                    | Wait for a cluster or project to become active.                                      |
| `kubectl`                 | Run your local `kubectl` against the cluster in the current context.                 |
| `token`, `auth get-token` | kubectl exec credential plugin (see [below](#kubectl-integration)).                  |
| `completion`              | Print a shell completion script for bash, zsh, fish or PowerShell.                   |

`ssh` and `clusters add-node` work only with nodes created through
docker-machine (RKE1). `up` applies a compose file. These commands are kept for
existing users.

## kubectl integration

When the Rancher setting `kubeconfig-generate-token` is `false`, kubeconfigs
downloaded from Rancher contain no token. Instead, `kubectl` runs this CLI as an
[exec credential plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins)
to get one:

- older kubeconfigs call `rancher token --server <url> --user <name>`
- newer kubeconfigs call
  `rancher auth get-token --server <url> --user-id <id> [--auth-provider <name>]`

The CLI must be on `PATH` for `kubectl` to find it. On first use it asks you
to log in through the auth provider and caches the token in the config file.
Later calls reuse the token until it expires.

- `--auth-provider` takes the name of an auth config, such as `github`,
  `local` or one of several OIDC or SAML configs. A provider name such as
  `genericoidc` also works. If it matches several configs, the CLI asks which
  one to use.
- For OAuth providers, `--auth-flow devicecode` (default) prints a URL and a
  code to enter in a browser. `--auth-flow authcode` opens a browser and
  receives the result on a local port.
- `--cacerts <file>` and `--skip-verify` control TLS verification of the
  server.
- `rancher token delete <name>|all` removes cached tokens.

## Configuration

`rancher login` writes the servers, the current context, and API and cached
kubectl tokens to `~/.rancher/cli2.json`. Change the directory with
`--config <dir>` or `RANCHER_CONFIG_DIR`.

> [!WARNING]
> The file holds API tokens in plaintext. The CLI creates it with mode `0600`
> in a `0700` directory. Don't copy it to shared locations, and delete servers
> you no longer use with `rancher server delete`.

Two per-server settings have no flag and are set by editing the file:

| Field                | Description                                        |
|----------------------|----------------------------------------------------|
| `proxyUrl`           | HTTP(S) proxy for requests to that server.         |
| `httpTimeoutSeconds` | Request timeout. Default is 60 seconds.            |

`--debug` turns on debug logging.

## Versioning

Rancher CLI is released in lockstep with Rancher: CLI `v2.X.Y` ships alongside
Rancher `v2.X.Y`, and the `v2.X` branch follows Rancher `v2.X`. `main` follows
the next Rancher minor. Release candidates are tagged the same way
(`v2.16.0-rc.4`).

Use the same minor version as your Rancher server where you can. Common
commands also work across versions, because the CLI and the server are often
upgraded at different times.

## Using it alongside rancher-cli

- Both binaries are named `rancher`. The first one on `PATH` wins; the
  rancher-cli installer uses `~/.local/bin`. Rename one of them to keep both.
- rancher-cli doesn't read or import `cli2.json`. Log in again with
  `rancher auth login`.
- Kubeconfigs that call `rancher token` or `rancher auth get-token` keep
  working after you switch binaries.
- rancher-cli doesn't support `ssh`, `up`, `clusters add-node` or
  `clusters export`. Keep this CLI if you need them.

## Development

Requires Go (version in [`go.mod`](go.mod)) and Docker for the image. Each file
in [`scripts/`](scripts) is a `make` target:

```
make                 # build, test, validate, package (the CI run)
make build           # ./bin/rancher for this machine
CROSS=1 make build   # also every release platform, into ./build/bin
make test            # go test with the race detector
make lint            # golangci-lint
make validate        # go mod tidy and verify; fails if the tree changes
make package         # archives of ./build/bin into ./dist, and a local rancher/cli image
```

`REPO=<name> make package` tags the image `<name>/cli` instead. Releases are
built by the [release workflow](.github/workflows/release.yml) from `v*` tags.

## Contact

For bugs, questions, comments, corrections and suggestions, open an issue in
[rancher/rancher](https://github.com/rancher/rancher/issues) with a title
starting with `[cli] `, or
[click here](https://github.com/rancher/rancher/issues/new?title=%5Bcli%5D%20).

Report security issues as described in the
[Rancher security policy](https://github.com/rancher/rancher/security/policy).

## License

Copyright (c) 2014-2026 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
