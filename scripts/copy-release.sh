#!/bin/bash
gsutil -m cp -r dist/artifacts/v*  gs://releases.rancher.com/cli
