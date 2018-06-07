package pki

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/rancher/rke/hosts"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"k8s.io/client-go/util/cert"
)

func GenerateSignedCertAndKey(
	caCrt *x509.Certificate,
	caKey *rsa.PrivateKey,
	serverCrt bool,
	commonName string,
	altNames *cert.AltNames,
	reusedKey *rsa.PrivateKey,
	orgs []string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Generate a generic signed certificate
	var rootKey *rsa.PrivateKey
	var err error
	rootKey = reusedKey
	if reusedKey == nil {
		rootKey, err = cert.NewPrivateKey()
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to generate private key for %s certificate: %v", commonName, err)
		}
	}
	usages := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	if serverCrt {
		usages = append(usages, x509.ExtKeyUsageServerAuth)
	}
	if altNames == nil {
		altNames = &cert.AltNames{}
	}
	caConfig := cert.Config{
		CommonName:   commonName,
		Organization: orgs,
		Usages:       usages,
		AltNames:     *altNames,
	}
	clientCert, err := cert.NewSignedCert(caConfig, rootKey, caCrt, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate %s certificate: %v", commonName, err)
	}
	return clientCert, rootKey, nil
}

func generateCACertAndKey() (*x509.Certificate, *rsa.PrivateKey, error) {
	rootKey, err := cert.NewPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate private key for CA certificate: %v", err)
	}
	caConfig := cert.Config{
		CommonName: CACertName,
	}
	kubeCACert, err := cert.NewSelfSignedCACert(caConfig, rootKey)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate CA certificate: %v", err)
	}

	return kubeCACert, rootKey, nil
}

func GetAltNames(cpHosts []*hosts.Host, clusterDomain string, KubernetesServiceIP net.IP, SANs []string) *cert.AltNames {
	ips := []net.IP{}
	dnsNames := []string{}
	for _, host := range cpHosts {
		// Check if node address is a valid IP
		if nodeIP := net.ParseIP(host.Address); nodeIP != nil {
			ips = append(ips, nodeIP)
		} else {
			dnsNames = append(dnsNames, host.Address)
		}

		// Check if node internal address is a valid IP
		if len(host.InternalAddress) != 0 && host.InternalAddress != host.Address {
			if internalIP := net.ParseIP(host.InternalAddress); internalIP != nil {
				ips = append(ips, internalIP)
			} else {
				dnsNames = append(dnsNames, host.InternalAddress)
			}
		}
		// Add hostname to the ALT dns names
		if len(host.HostnameOverride) != 0 && host.HostnameOverride != host.Address {
			dnsNames = append(dnsNames, host.HostnameOverride)
		}
	}

	for _, host := range SANs {
		// Check if node address is a valid IP
		if nodeIP := net.ParseIP(host); nodeIP != nil {
			ips = append(ips, nodeIP)
		} else {
			dnsNames = append(dnsNames, host)
		}
	}

	ips = append(ips, net.ParseIP("127.0.0.1"))
	ips = append(ips, KubernetesServiceIP)
	dnsNames = append(dnsNames, []string{
		"localhost",
		"kubernetes",
		"kubernetes.default",
		"kubernetes.default.svc",
		"kubernetes.default.svc." + clusterDomain,
	}...)
	return &cert.AltNames{
		IPs:      ips,
		DNSNames: dnsNames,
	}
}

func (c *CertificatePKI) ToEnv() []string {
	env := []string{}
	if c.Key != nil {
		env = append(env, c.KeyToEnv())
	}
	if c.Certificate != nil {
		env = append(env, c.CertToEnv())
	}
	if c.Config != "" && c.ConfigEnvName != "" {
		env = append(env, c.ConfigToEnv())
	}
	return env
}

func (c *CertificatePKI) CertToEnv() string {
	encodedCrt := cert.EncodeCertPEM(c.Certificate)
	return fmt.Sprintf("%s=%s", c.EnvName, string(encodedCrt))
}

func (c *CertificatePKI) KeyToEnv() string {
	encodedKey := cert.EncodePrivateKeyPEM(c.Key)
	return fmt.Sprintf("%s=%s", c.KeyEnvName, string(encodedKey))
}

func (c *CertificatePKI) ConfigToEnv() string {
	return fmt.Sprintf("%s=%s", c.ConfigEnvName, c.Config)
}

func getEnvFromName(name string) string {
	return strings.Replace(strings.ToUpper(name), "-", "_", -1)
}

func getKeyEnvFromEnv(env string) string {
	return fmt.Sprintf("%s_KEY", env)
}

func getConfigEnvFromEnv(env string) string {
	return fmt.Sprintf("KUBECFG_%s", env)
}

func GetEtcdCrtName(address string) string {
	newAddress := strings.Replace(address, ".", "-", -1)
	return fmt.Sprintf("%s-%s", EtcdCertName, newAddress)
}

func GetCertPath(name string) string {
	return fmt.Sprintf("%s%s.pem", CertPathPrefix, name)
}

func GetKeyPath(name string) string {
	return fmt.Sprintf("%s%s-key.pem", CertPathPrefix, name)
}

func GetConfigPath(name string) string {
	return fmt.Sprintf("%skubecfg-%s.yaml", CertPathPrefix, name)
}

func GetCertTempPath(name string) string {
	return fmt.Sprintf("%s%s.pem", TempCertPath, name)
}

func GetKeyTempPath(name string) string {
	return fmt.Sprintf("%s%s-key.pem", TempCertPath, name)
}

func GetConfigTempPath(name string) string {
	return fmt.Sprintf("%skubecfg-%s.yaml", TempCertPath, name)
}

func ToCertObject(componentName, commonName, ouName string, cert *x509.Certificate, key *rsa.PrivateKey) CertificatePKI {
	var config, configPath, configEnvName string
	if len(commonName) == 0 {
		commonName = getDefaultCN(componentName)
	}

	envName := getEnvFromName(componentName)
	keyEnvName := getKeyEnvFromEnv(envName)
	caCertPath := GetCertPath(CACertName)
	path := GetCertPath(componentName)
	keyPath := GetKeyPath(componentName)

	if componentName != CACertName && componentName != KubeAPICertName && !strings.Contains(componentName, EtcdCertName) {
		config = getKubeConfigX509("https://127.0.0.1:6443", "local", componentName, caCertPath, path, keyPath)
		configPath = GetConfigPath(componentName)
		configEnvName = getConfigEnvFromEnv(envName)
	}

	return CertificatePKI{
		Certificate:   cert,
		Key:           key,
		Config:        config,
		Name:          componentName,
		CommonName:    commonName,
		OUName:        ouName,
		EnvName:       envName,
		KeyEnvName:    keyEnvName,
		ConfigEnvName: configEnvName,
		Path:          path,
		KeyPath:       keyPath,
		ConfigPath:    configPath,
	}
}

func getDefaultCN(name string) string {
	return fmt.Sprintf("system:%s", name)
}

func getControlCertKeys() []string {
	return []string{
		CACertName,
		KubeAPICertName,
		KubeControllerCertName,
		KubeSchedulerCertName,
		KubeProxyCertName,
		KubeNodeCertName,
		EtcdClientCertName,
		EtcdClientCACertName,
	}
}

func getWorkerCertKeys() []string {
	return []string{
		CACertName,
		KubeProxyCertName,
		KubeNodeCertName,
	}
}

func getEtcdCertKeys(rkeNodes []v3.RKEConfigNode, etcdRole string) []string {
	certList := []string{
		CACertName,
		KubeProxyCertName,
		KubeNodeCertName,
	}
	etcdHosts := hosts.NodesToHosts(rkeNodes, etcdRole)
	for _, host := range etcdHosts {
		certList = append(certList, GetEtcdCrtName(host.InternalAddress))
	}
	return certList

}

func GetKubernetesServiceIP(serviceClusterRange string) (net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(serviceClusterRange)
	if err != nil {
		return nil, fmt.Errorf("Failed to get kubernetes service IP from Kube API option [service_cluster_ip_range]: %v", err)
	}
	ip = ip.Mask(ipnet.Mask)
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
	return ip, nil
}

func GetLocalKubeConfig(configPath, configDir string) string {
	baseDir := filepath.Dir(configPath)
	if len(configDir) > 0 {
		baseDir = filepath.Dir(configDir)
	}
	fileName := filepath.Base(configPath)
	baseDir += "/"
	return fmt.Sprintf("%s%s%s", baseDir, KubeAdminConfigPrefix, fileName)
}

func strCrtToEnv(crtName, crt string) string {
	return fmt.Sprintf("%s=%s", getEnvFromName(crtName), crt)
}

func strKeyToEnv(crtName, key string) string {
	envName := getEnvFromName(crtName)
	return fmt.Sprintf("%s=%s", getKeyEnvFromEnv(envName), key)
}
