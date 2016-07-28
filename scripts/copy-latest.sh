#!/bin/bash
gsutil -m rsync -r dist/artifacts/latest/   gs://releases.rancher.com/cli/latest
