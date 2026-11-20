# Contributing

We are not accepting code contributions at the moment. A clear issue is the
most useful thing you can send.

Issues are turned off in this repository. Open them in
[rancher/rancher](https://github.com/rancher/rancher/issues) with a title that
starts with `[cli] `.

## Report a bug

Use the [bug report template](https://github.com/rancher/rancher/issues/new?template=bug_report.md&title=%5Bcli%5D%20)
and include:

- the CLI version (`rancher --version`)
- the Rancher version, and whether it is community or Prime
- your operating system, architecture and shell
- the exact command you ran
- the full output, if the command printed an error: run the command again
  with `--debug` and paste the whole output
- what you expected to happen
- steps that reproduce it from a clean config directory:
  `RANCHER_CONFIG_DIR=$(mktemp -d) rancher login ...`
  (on Windows, set `RANCHER_CONFIG_DIR` to a new empty directory)

Remove tokens, passwords and server URLs that must stay private from
everything you paste.

## Request a feature

Use the [feature request template](https://github.com/rancher/rancher/issues/new?template=feature_request.md&title=%5Bcli%5D%20).
Describe the workflow you are trying to do and how you do it today, not
only the flag or command you want.

## Report a security issue

Never in a public issue: see the
[Rancher security policy](https://github.com/rancher/rancher/security/policy).

## Ask a question

Use the [Rancher forums](https://forums.rancher.com/) or
[Rancher Slack](https://slack.rancher.io/).
