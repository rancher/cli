package model

import "github.com/rancher/go-rancher/client"

//UpgradeInfo structure contains the new version info
type UpgradeInfo struct {
	client.Resource
	CurrentVersion  string            `json:"currentVersion"`
	NewVersionLinks map[string]string `json:"newVersionLinks"`
}
