package pki

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/rke/log"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"k8s.io/client-go/util/cert"
)

type CertificatePKI struct {
	Certificate   *x509.Certificate
	Key           *rsa.PrivateKey
	Config        string
	Name          string
	CommonName    string
	OUName        string
	EnvName       string
	Path          string
	KeyEnvName    string
	KeyPath       string
	ConfigEnvName string
	ConfigPath    string
}

const (
	etcdRole    = "etcd"
	controlRole = "controlplane"
	workerRole  = "worker"
)

func GenerateRKECerts(ctx context.Context, rkeConfig v3.RancherKubernetesEngineConfig, configPath, configDir string) (map[string]CertificatePKI, error) {
	certs := make(map[string]CertificatePKI)
	// generate CA certificate and key
	log.Infof(ctx, "[certificates] Generating CA kubernetes certificates")
	caCrt, caKey, err := generateCACertAndKey()
	if err != nil {
		return nil, err
	}
	certs[CACertName] = ToCertObject(CACertName, "", "", caCrt, caKey)

	// generate API certificate and key
	log.Infof(ctx, "[certificates] Generating Kubernetes API server certificates")
	kubernetesServiceIP, err := GetKubernetesServiceIP(rkeConfig.Services.KubeAPI.ServiceClusterIPRange)
	clusterDomain := rkeConfig.Services.Kubelet.ClusterDomain
	cpHosts := hosts.NodesToHosts(rkeConfig.Nodes, controlRole)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Kubernetes Service IP: %v", err)
	}
	kubeAPIAltNames := GetAltNames(cpHosts, clusterDomain, kubernetesServiceIP, rkeConfig.Authentication.SANs)
	kubeAPICrt, kubeAPIKey, err := GenerateSignedCertAndKey(caCrt, caKey, true, KubeAPICertName, kubeAPIAltNames, nil, nil)
	if err != nil {
		return nil, err
	}
	certs[KubeAPICertName] = ToCertObject(KubeAPICertName, "", "", kubeAPICrt, kubeAPIKey)

	// generate Kube controller-manager certificate and key
	log.Infof(ctx, "[certificates] Generating Kube Controller certificates")
	kubeControllerCrt, kubeControllerKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, getDefaultCN(KubeControllerCertName), nil, nil, nil)
	if err != nil {
		return nil, err
	}
	certs[KubeControllerCertName] = ToCertObject(KubeControllerCertName, "", "", kubeControllerCrt, kubeControllerKey)

	// generate Kube scheduler certificate and key
	log.Infof(ctx, "[certificates] Generating Kube Scheduler certificates")
	kubeSchedulerCrt, kubeSchedulerKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, getDefaultCN(KubeSchedulerCertName), nil, nil, nil)
	if err != nil {
		return nil, err
	}
	certs[KubeSchedulerCertName] = ToCertObject(KubeSchedulerCertName, "", "", kubeSchedulerCrt, kubeSchedulerKey)

	// generate Kube Proxy certificate and key
	log.Infof(ctx, "[certificates] Generating Kube Proxy certificates")
	kubeProxyCrt, kubeProxyKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, getDefaultCN(KubeProxyCertName), nil, nil, nil)
	if err != nil {
		return nil, err
	}
	certs[KubeProxyCertName] = ToCertObject(KubeProxyCertName, "", "", kubeProxyCrt, kubeProxyKey)

	// generate Kubelet certificate and key
	log.Infof(ctx, "[certificates] Generating Node certificate")
	nodeCrt, nodeKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, KubeNodeCommonName, nil, nil, []string{KubeNodeOrganizationName})
	if err != nil {
		return nil, err
	}
	certs[KubeNodeCertName] = ToCertObject(KubeNodeCertName, KubeNodeCommonName, KubeNodeOrganizationName, nodeCrt, nodeKey)

	// generate Admin certificate and key
	log.Infof(ctx, "[certificates] Generating admin certificates and kubeconfig")
	if len(configPath) == 0 {
		configPath = ClusterConfig
	}
	localKubeConfigPath := GetLocalKubeConfig(configPath, configDir)
	kubeAdminCrt, kubeAdminKey, err := GenerateSignedCertAndKey(caCrt, caKey, false, KubeAdminCertName, nil, nil, []string{KubeAdminOrganizationName})
	if err != nil {
		return nil, err
	}
	kubeAdminCertObj := ToCertObject(KubeAdminCertName, KubeAdminCertName, KubeAdminOrganizationName, kubeAdminCrt, kubeAdminKey)
	if len(cpHosts) > 0 {
		kubeAdminConfig := GetKubeConfigX509WithData(
			"https://"+cpHosts[0].Address+":6443",
			rkeConfig.ClusterName,
			KubeAdminCertName,
			string(cert.EncodeCertPEM(caCrt)),
			string(cert.EncodeCertPEM(kubeAdminCrt)),
			string(cert.EncodePrivateKeyPEM(kubeAdminKey)))
		kubeAdminCertObj.Config = kubeAdminConfig
		kubeAdminCertObj.ConfigPath = localKubeConfigPath
	} else {
		kubeAdminCertObj.Config = ""
	}
	certs[KubeAdminCertName] = kubeAdminCertObj
	// generate etcd certificate and key
	if len(rkeConfig.Services.Etcd.ExternalURLs) > 0 {
		clientCert, err := cert.ParseCertsPEM([]byte(rkeConfig.Services.Etcd.Cert))
		if err != nil {
			return nil, err
		}
		clientKey, err := cert.ParsePrivateKeyPEM([]byte(rkeConfig.Services.Etcd.Key))
		if err != nil {
			return nil, err
		}
		certs[EtcdClientCertName] = ToCertObject(EtcdClientCertName, "", "", clientCert[0], clientKey.(*rsa.PrivateKey))

		caCert, err := cert.ParseCertsPEM([]byte(rkeConfig.Services.Etcd.CACert))
		if err != nil {
			return nil, err
		}
		certs[EtcdClientCACertName] = ToCertObject(EtcdClientCACertName, "", "", caCert[0], nil)
	}
	etcdHosts := hosts.NodesToHosts(rkeConfig.Nodes, etcdRole)
	etcdAltNames := GetAltNames(etcdHosts, clusterDomain, kubernetesServiceIP, []string{})
	for _, host := range etcdHosts {
		log.Infof(ctx, "[certificates] Generating etcd-%s certificate and key", host.InternalAddress)
		etcdCrt, etcdKey, err := GenerateSignedCertAndKey(caCrt, caKey, true, EtcdCertName, etcdAltNames, nil, nil)
		if err != nil {
			return nil, err
		}
		etcdName := GetEtcdCrtName(host.InternalAddress)
		certs[etcdName] = ToCertObject(etcdName, "", "", etcdCrt, etcdKey)
	}

	return certs, nil
}

func GenerateRKENodeCerts(ctx context.Context, rkeConfig v3.RancherKubernetesEngineConfig, nodeAddress string, certBundle map[string]CertificatePKI) map[string]CertificatePKI {
	crtMap := make(map[string]CertificatePKI)
	crtKeys := []string{}
	removeCAKey := true
	for _, node := range rkeConfig.Nodes {
		if node.Address == nodeAddress {
			for _, role := range node.Role {
				switch role {
				case controlRole:
					keys := getControlCertKeys()
					crtKeys = append(crtKeys, keys...)
					removeCAKey = false
				case workerRole:
					keys := getWorkerCertKeys()
					crtKeys = append(crtKeys, keys...)
				case etcdRole:
					keys := getEtcdCertKeys(rkeConfig.Nodes, etcdRole)
					crtKeys = append(crtKeys, keys...)
				}
			}
			break
		}
	}
	for _, key := range crtKeys {
		crtMap[key] = certBundle[key]
	}
	if removeCAKey {
		caCert := crtMap[CACertName]
		caCert.Key = nil
		caCert.KeyEnvName = ""
		caCert.KeyPath = ""
		crtMap[CACertName] = caCert
	}
	return crtMap
}

func RegenerateEtcdCertificate(
	ctx context.Context,
	crtMap map[string]CertificatePKI,
	etcdHost *hosts.Host,
	etcdHosts []*hosts.Host,
	clusterDomain string,
	KubernetesServiceIP net.IP) (map[string]CertificatePKI, error) {

	log.Infof(ctx, "[certificates] Regenerating new etcd-%s certificate and key", etcdHost.InternalAddress)
	caCrt := crtMap[CACertName].Certificate
	caKey := crtMap[CACertName].Key
	etcdAltNames := GetAltNames(etcdHosts, clusterDomain, KubernetesServiceIP, []string{})

	etcdCrt, etcdKey, err := GenerateSignedCertAndKey(caCrt, caKey, true, EtcdCertName, etcdAltNames, nil, nil)
	if err != nil {
		return nil, err
	}
	etcdName := GetEtcdCrtName(etcdHost.InternalAddress)
	crtMap[etcdName] = ToCertObject(etcdName, "", "", etcdCrt, etcdKey)
	log.Infof(ctx, "[certificates] Successfully generated new etcd-%s certificate and key", etcdHost.InternalAddress)
	return crtMap, nil
}
