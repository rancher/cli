package client

const (
	ClusterPipelineSpecType              = "clusterPipelineSpec"
	ClusterPipelineSpecFieldClusterId    = "clusterId"
	ClusterPipelineSpecFieldDeploy       = "deploy"
	ClusterPipelineSpecFieldGithubConfig = "githubConfig"
)

type ClusterPipelineSpec struct {
	ClusterId    string               `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	Deploy       bool                 `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	GithubConfig *GithubClusterConfig `json:"githubConfig,omitempty" yaml:"githubConfig,omitempty"`
}
