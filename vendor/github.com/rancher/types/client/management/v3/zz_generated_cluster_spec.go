package client

const (
	ClusterSpecType                                     = "clusterSpec"
	ClusterSpecFieldAzureKubernetesServiceConfig        = "azureKubernetesServiceConfig"
	ClusterSpecFieldDefaultClusterRoleForProjectMembers = "defaultClusterRoleForProjectMembers"
	ClusterSpecFieldDefaultPodSecurityPolicyTemplateId  = "defaultPodSecurityPolicyTemplateId"
	ClusterSpecFieldDescription                         = "description"
	ClusterSpecFieldDisplayName                         = "displayName"
	ClusterSpecFieldGoogleKubernetesEngineConfig        = "googleKubernetesEngineConfig"
	ClusterSpecFieldImportedConfig                      = "importedConfig"
	ClusterSpecFieldInternal                            = "internal"
	ClusterSpecFieldRancherKubernetesEngineConfig       = "rancherKubernetesEngineConfig"
)

type ClusterSpec struct {
	AzureKubernetesServiceConfig        *AzureKubernetesServiceConfig  `json:"azureKubernetesServiceConfig,omitempty"`
	DefaultClusterRoleForProjectMembers string                         `json:"defaultClusterRoleForProjectMembers,omitempty"`
	DefaultPodSecurityPolicyTemplateId  string                         `json:"defaultPodSecurityPolicyTemplateId,omitempty"`
	Description                         string                         `json:"description,omitempty"`
	DisplayName                         string                         `json:"displayName,omitempty"`
	GoogleKubernetesEngineConfig        *GoogleKubernetesEngineConfig  `json:"googleKubernetesEngineConfig,omitempty"`
	ImportedConfig                      *ImportedConfig                `json:"importedConfig,omitempty"`
	Internal                            bool                           `json:"internal,omitempty"`
	RancherKubernetesEngineConfig       *RancherKubernetesEngineConfig `json:"rancherKubernetesEngineConfig,omitempty"`
}
