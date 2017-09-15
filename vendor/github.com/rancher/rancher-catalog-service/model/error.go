package model

import "github.com/rancher/go-rancher/client"

//CatalogError structure contains the error resource definition
type CatalogError struct {
	client.Resource
	Status  string `json:"status"`
	Message string `json:"message"`
}
