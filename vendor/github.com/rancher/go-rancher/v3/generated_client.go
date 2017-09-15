package client

type RancherClient struct {
	RancherBaseClient

	Account                            AccountOperations
	AddOutputsInput                    AddOutputsInputOperations
	Agent                              AgentOperations
	Amazonec2Config                    Amazonec2ConfigOperations
	ApiKey                             ApiKeyOperations
	AuditLog                           AuditLogOperations
	AzureConfig                        AzureConfigOperations
	Azureadconfig                      AzureadconfigOperations
	BaseMachineConfig                  BaseMachineConfigOperations
	BlkioDeviceOption                  BlkioDeviceOptionOperations
	Certificate                        CertificateOperations
	ChangeSecretInput                  ChangeSecretInputOperations
	Cluster                            ClusterOperations
	ClusterIdentity                    ClusterIdentityOperations
	ComposeConfig                      ComposeConfigOperations
	ComposeConfigInput                 ComposeConfigInputOperations
	Container                          ContainerOperations
	ContainerConfig                    ContainerConfigOperations
	ContainerEvent                     ContainerEventOperations
	ContainerExec                      ContainerExecOperations
	ContainerLogs                      ContainerLogsOperations
	ContainerProxy                     ContainerProxyOperations
	ContainerUpgrade                   ContainerUpgradeOperations
	Credential                         CredentialOperations
	Databasechangelog                  DatabasechangelogOperations
	Databasechangeloglock              DatabasechangeloglockOperations
	DefaultNetwork                     DefaultNetworkOperations
	DependsOn                          DependsOnOperations
	DeploymentSyncRequest              DeploymentSyncRequestOperations
	DeploymentSyncResponse             DeploymentSyncResponseOperations
	DeploymentUnit                     DeploymentUnitOperations
	DigitaloceanConfig                 DigitaloceanConfigOperations
	DnsService                         DnsServiceOperations
	DynamicSchema                      DynamicSchemaOperations
	EnvironmentInfo                    EnvironmentInfoOperations
	Error                              ErrorOperations
	ExternalDnsEvent                   ExternalDnsEventOperations
	ExternalEvent                      ExternalEventOperations
	ExternalHostEvent                  ExternalHostEventOperations
	ExternalService                    ExternalServiceOperations
	ExternalServiceEvent               ExternalServiceEventOperations
	FieldDocumentation                 FieldDocumentationOperations
	GenericObject                      GenericObjectOperations
	HaMembership                       HaMembershipOperations
	HealthcheckInfo                    HealthcheckInfoOperations
	HealthcheckState                   HealthcheckStateOperations
	Host                               HostOperations
	HostAccess                         HostAccessOperations
	HostApiProxyToken                  HostApiProxyTokenOperations
	HostInfo                           HostInfoOperations
	HostTemplate                       HostTemplateOperations
	Identity                           IdentityOperations
	InServiceUpgradeStrategy           InServiceUpgradeStrategyOperations
	Instance                           InstanceOperations
	InstanceConsole                    InstanceConsoleOperations
	InstanceConsoleInput               InstanceConsoleInputOperations
	InstanceHealthCheck                InstanceHealthCheckOperations
	InstanceInfo                       InstanceInfoOperations
	InstanceRemove                     InstanceRemoveOperations
	InstanceStatus                     InstanceStatusOperations
	InstanceStop                       InstanceStopOperations
	K8sClientConfig                    K8sClientConfigOperations
	K8sServerConfig                    K8sServerConfigOperations
	LaunchConfig                       LaunchConfigOperations
	LbConfig                           LbConfigOperations
	LbTargetConfig                     LbTargetConfigOperations
	Ldapconfig                         LdapconfigOperations
	Link                               LinkOperations
	LoadBalancerCookieStickinessPolicy LoadBalancerCookieStickinessPolicyOperations
	LoadBalancerService                LoadBalancerServiceOperations
	LocalAuthConfig                    LocalAuthConfigOperations
	LogConfig                          LogConfigOperations
	MachineDriver                      MachineDriverOperations
	MetadataObject                     MetadataObjectOperations
	MetadataSyncRequest                MetadataSyncRequestOperations
	Mount                              MountOperations
	MountEntry                         MountEntryOperations
	Network                            NetworkOperations
	NetworkDriver                      NetworkDriverOperations
	NetworkDriverService               NetworkDriverServiceOperations
	NetworkInfo                        NetworkInfoOperations
	NetworkPolicyRule                  NetworkPolicyRuleOperations
	NetworkPolicyRuleBetween           NetworkPolicyRuleBetweenOperations
	NetworkPolicyRuleMember            NetworkPolicyRuleMemberOperations
	NetworkPolicyRuleWithin            NetworkPolicyRuleWithinOperations
	Openldapconfig                     OpenldapconfigOperations
	PacketConfig                       PacketConfigOperations
	Password                           PasswordOperations
	PortRule                           PortRuleOperations
	ProcessExecution                   ProcessExecutionOperations
	ProcessInstance                    ProcessInstanceOperations
	ProcessPool                        ProcessPoolOperations
	ProcessSummary                     ProcessSummaryOperations
	Project                            ProjectOperations
	ProjectMember                      ProjectMemberOperations
	PublicEndpoint                     PublicEndpointOperations
	Publish                            PublishOperations
	PullTask                           PullTaskOperations
	Register                           RegisterOperations
	RegistrationToken                  RegistrationTokenOperations
	Registry                           RegistryOperations
	RegistryCredential                 RegistryCredentialOperations
	RestartPolicy                      RestartPolicyOperations
	Revision                           RevisionOperations
	ScalingGroup                       ScalingGroupOperations
	ScheduledUpgrade                   ScheduledUpgradeOperations
	Secret                             SecretOperations
	SecretReference                    SecretReferenceOperations
	SelectorService                    SelectorServiceOperations
	Service                            ServiceOperations
	ServiceEvent                       ServiceEventOperations
	ServiceInfo                        ServiceInfoOperations
	ServiceLog                         ServiceLogOperations
	ServiceProxy                       ServiceProxyOperations
	ServiceRollback                    ServiceRollbackOperations
	ServiceUpgrade                     ServiceUpgradeOperations
	ServiceUpgradeStrategy             ServiceUpgradeStrategyOperations
	ServicesPortRange                  ServicesPortRangeOperations
	SetProjectMembersInput             SetProjectMembersInputOperations
	Setting                            SettingOperations
	Stack                              StackOperations
	StackConfiguration                 StackConfigurationOperations
	StackInfo                          StackInfoOperations
	StackUpgrade                       StackUpgradeOperations
	StatsAccess                        StatsAccessOperations
	StorageDriver                      StorageDriverOperations
	StorageDriverService               StorageDriverServiceOperations
	StoragePool                        StoragePoolOperations
	Subnet                             SubnetOperations
	Subscribe                          SubscribeOperations
	TargetPortRule                     TargetPortRuleOperations
	TypeDocumentation                  TypeDocumentationOperations
	Ulimit                             UlimitOperations
	VirtualMachine                     VirtualMachineOperations
	VirtualMachineDisk                 VirtualMachineDiskOperations
	Volume                             VolumeOperations
	VolumeActivateInput                VolumeActivateInputOperations
	VolumeTemplate                     VolumeTemplateOperations
}

func constructClient(rancherBaseClient *RancherBaseClientImpl) *RancherClient {
	client := &RancherClient{
		RancherBaseClient: rancherBaseClient,
	}

	client.Account = newAccountClient(client)
	client.AddOutputsInput = newAddOutputsInputClient(client)
	client.Agent = newAgentClient(client)
	client.Amazonec2Config = newAmazonec2ConfigClient(client)
	client.ApiKey = newApiKeyClient(client)
	client.AuditLog = newAuditLogClient(client)
	client.AzureConfig = newAzureConfigClient(client)
	client.Azureadconfig = newAzureadconfigClient(client)
	client.BaseMachineConfig = newBaseMachineConfigClient(client)
	client.BlkioDeviceOption = newBlkioDeviceOptionClient(client)
	client.Certificate = newCertificateClient(client)
	client.ChangeSecretInput = newChangeSecretInputClient(client)
	client.Cluster = newClusterClient(client)
	client.ClusterIdentity = newClusterIdentityClient(client)
	client.ComposeConfig = newComposeConfigClient(client)
	client.ComposeConfigInput = newComposeConfigInputClient(client)
	client.Container = newContainerClient(client)
	client.ContainerConfig = newContainerConfigClient(client)
	client.ContainerEvent = newContainerEventClient(client)
	client.ContainerExec = newContainerExecClient(client)
	client.ContainerLogs = newContainerLogsClient(client)
	client.ContainerProxy = newContainerProxyClient(client)
	client.ContainerUpgrade = newContainerUpgradeClient(client)
	client.Credential = newCredentialClient(client)
	client.Databasechangelog = newDatabasechangelogClient(client)
	client.Databasechangeloglock = newDatabasechangeloglockClient(client)
	client.DefaultNetwork = newDefaultNetworkClient(client)
	client.DependsOn = newDependsOnClient(client)
	client.DeploymentSyncRequest = newDeploymentSyncRequestClient(client)
	client.DeploymentSyncResponse = newDeploymentSyncResponseClient(client)
	client.DeploymentUnit = newDeploymentUnitClient(client)
	client.DigitaloceanConfig = newDigitaloceanConfigClient(client)
	client.DnsService = newDnsServiceClient(client)
	client.DynamicSchema = newDynamicSchemaClient(client)
	client.EnvironmentInfo = newEnvironmentInfoClient(client)
	client.Error = newErrorClient(client)
	client.ExternalDnsEvent = newExternalDnsEventClient(client)
	client.ExternalEvent = newExternalEventClient(client)
	client.ExternalHostEvent = newExternalHostEventClient(client)
	client.ExternalService = newExternalServiceClient(client)
	client.ExternalServiceEvent = newExternalServiceEventClient(client)
	client.FieldDocumentation = newFieldDocumentationClient(client)
	client.GenericObject = newGenericObjectClient(client)
	client.HaMembership = newHaMembershipClient(client)
	client.HealthcheckInfo = newHealthcheckInfoClient(client)
	client.HealthcheckState = newHealthcheckStateClient(client)
	client.Host = newHostClient(client)
	client.HostAccess = newHostAccessClient(client)
	client.HostApiProxyToken = newHostApiProxyTokenClient(client)
	client.HostInfo = newHostInfoClient(client)
	client.HostTemplate = newHostTemplateClient(client)
	client.Identity = newIdentityClient(client)
	client.InServiceUpgradeStrategy = newInServiceUpgradeStrategyClient(client)
	client.Instance = newInstanceClient(client)
	client.InstanceConsole = newInstanceConsoleClient(client)
	client.InstanceConsoleInput = newInstanceConsoleInputClient(client)
	client.InstanceHealthCheck = newInstanceHealthCheckClient(client)
	client.InstanceInfo = newInstanceInfoClient(client)
	client.InstanceRemove = newInstanceRemoveClient(client)
	client.InstanceStatus = newInstanceStatusClient(client)
	client.InstanceStop = newInstanceStopClient(client)
	client.K8sClientConfig = newK8sClientConfigClient(client)
	client.K8sServerConfig = newK8sServerConfigClient(client)
	client.LaunchConfig = newLaunchConfigClient(client)
	client.LbConfig = newLbConfigClient(client)
	client.LbTargetConfig = newLbTargetConfigClient(client)
	client.Ldapconfig = newLdapconfigClient(client)
	client.Link = newLinkClient(client)
	client.LoadBalancerCookieStickinessPolicy = newLoadBalancerCookieStickinessPolicyClient(client)
	client.LoadBalancerService = newLoadBalancerServiceClient(client)
	client.LocalAuthConfig = newLocalAuthConfigClient(client)
	client.LogConfig = newLogConfigClient(client)
	client.MachineDriver = newMachineDriverClient(client)
	client.MetadataObject = newMetadataObjectClient(client)
	client.MetadataSyncRequest = newMetadataSyncRequestClient(client)
	client.Mount = newMountClient(client)
	client.MountEntry = newMountEntryClient(client)
	client.Network = newNetworkClient(client)
	client.NetworkDriver = newNetworkDriverClient(client)
	client.NetworkDriverService = newNetworkDriverServiceClient(client)
	client.NetworkInfo = newNetworkInfoClient(client)
	client.NetworkPolicyRule = newNetworkPolicyRuleClient(client)
	client.NetworkPolicyRuleBetween = newNetworkPolicyRuleBetweenClient(client)
	client.NetworkPolicyRuleMember = newNetworkPolicyRuleMemberClient(client)
	client.NetworkPolicyRuleWithin = newNetworkPolicyRuleWithinClient(client)
	client.Openldapconfig = newOpenldapconfigClient(client)
	client.PacketConfig = newPacketConfigClient(client)
	client.Password = newPasswordClient(client)
	client.PortRule = newPortRuleClient(client)
	client.ProcessExecution = newProcessExecutionClient(client)
	client.ProcessInstance = newProcessInstanceClient(client)
	client.ProcessPool = newProcessPoolClient(client)
	client.ProcessSummary = newProcessSummaryClient(client)
	client.Project = newProjectClient(client)
	client.ProjectMember = newProjectMemberClient(client)
	client.PublicEndpoint = newPublicEndpointClient(client)
	client.Publish = newPublishClient(client)
	client.PullTask = newPullTaskClient(client)
	client.Register = newRegisterClient(client)
	client.RegistrationToken = newRegistrationTokenClient(client)
	client.Registry = newRegistryClient(client)
	client.RegistryCredential = newRegistryCredentialClient(client)
	client.RestartPolicy = newRestartPolicyClient(client)
	client.Revision = newRevisionClient(client)
	client.ScalingGroup = newScalingGroupClient(client)
	client.ScheduledUpgrade = newScheduledUpgradeClient(client)
	client.Secret = newSecretClient(client)
	client.SecretReference = newSecretReferenceClient(client)
	client.SelectorService = newSelectorServiceClient(client)
	client.Service = newServiceClient(client)
	client.ServiceEvent = newServiceEventClient(client)
	client.ServiceInfo = newServiceInfoClient(client)
	client.ServiceLog = newServiceLogClient(client)
	client.ServiceProxy = newServiceProxyClient(client)
	client.ServiceRollback = newServiceRollbackClient(client)
	client.ServiceUpgrade = newServiceUpgradeClient(client)
	client.ServiceUpgradeStrategy = newServiceUpgradeStrategyClient(client)
	client.ServicesPortRange = newServicesPortRangeClient(client)
	client.SetProjectMembersInput = newSetProjectMembersInputClient(client)
	client.Setting = newSettingClient(client)
	client.Stack = newStackClient(client)
	client.StackConfiguration = newStackConfigurationClient(client)
	client.StackInfo = newStackInfoClient(client)
	client.StackUpgrade = newStackUpgradeClient(client)
	client.StatsAccess = newStatsAccessClient(client)
	client.StorageDriver = newStorageDriverClient(client)
	client.StorageDriverService = newStorageDriverServiceClient(client)
	client.StoragePool = newStoragePoolClient(client)
	client.Subnet = newSubnetClient(client)
	client.Subscribe = newSubscribeClient(client)
	client.TargetPortRule = newTargetPortRuleClient(client)
	client.TypeDocumentation = newTypeDocumentationClient(client)
	client.Ulimit = newUlimitClient(client)
	client.VirtualMachine = newVirtualMachineClient(client)
	client.VirtualMachineDisk = newVirtualMachineDiskClient(client)
	client.Volume = newVolumeClient(client)
	client.VolumeActivateInput = newVolumeActivateInputClient(client)
	client.VolumeTemplate = newVolumeTemplateClient(client)

	return client
}

func NewRancherClient(opts *ClientOpts) (*RancherClient, error) {
	rancherBaseClient := &RancherBaseClientImpl{
		Types: map[string]Schema{},
	}
	client := constructClient(rancherBaseClient)

	err := setupRancherBaseClient(rancherBaseClient, opts)
	if err != nil {
		return nil, err
	}

	return client, nil
}
