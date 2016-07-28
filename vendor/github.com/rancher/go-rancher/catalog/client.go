package catalog

type RancherBaseClient struct {
	Opts    *ClientOpts
	Schemas *Schemas
	Types   map[string]Schema
}
