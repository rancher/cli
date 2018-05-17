package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/rancher/cli/templates"
	managementApi "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	//	"os"
	//	"os/exec"
	//	"strings"
)

const (
	CERTmanagerVer = "v0.3.0"
	CERTManagerURL = "https://raw.githubusercontent.com/jetstack/cert-manager/" + CERTmanagerVer + "/contrib/manifests/cert-manager/with-rbac.yaml"
	HAOutputRKE    = `rke`
	HAOutputK8S    = `k8s`
	HADescription  = `Create Rancher HA template and deployment manifest for rke and k8s.`
	AcmeStagingCA  = `https://letsencrypt.org/certs/fakelerootx1.pem`
	HATemplateDesc = `Create Rancher HA minimal template to be used by rancher ha up command or get k8s deployment manifest.

Output modes:
- ` + HAOutputRKE + `: Generate Rancher HA minimal template. Add rke nodes info and use it with rancher ha up command to deploy your rke cluster.
- ` + HAOutputK8S + `: Generate Rancher HA template and k8s deployment manifest. Use manifest with kubectl to deploy it in your k8s cluster.

TLS modes:
- ` + templates.TLSModeIngress + `: Ingress as SSL termination. Configure Rancher HA deployment to use user provided TLS certs. 
  tls-key and tls-cert are required. tls-ca is required if selfsigned certs are used. Certs must be valid for hostname param. 
- ` + templates.TLSModeRancher + `: Rancher as SSL termination. Configure Rancher HA deployment to auto generate TLS certs. Just for ` + HAOutputRKE + ` output-mode. 
- ` + templates.TLSModeExternal + `: External LB as SSL termination. Configure Rancher HA deployment don't use TLS certs.

TLS issuer for TLS mode ` + templates.TLSModeIngress + `:
- ` + templates.TLSIssuerAcme + `: Configure Rancher HA deployment to use letsencrypt TLS certs.
- ` + templates.TLSIssuerCA + `: Configure Rancher HA deployment to use CA issuer to manage TLS certs.
`
	RancherImage     = `rancher/rancher:latest`
	RKEEtcdLabel     = `etcd`
	RKEMinimalEtcd   = 3
	RKEMinimalVer    = `0.1.6`
	RKEHAMinimalTmpl = `# Minimal RKE Cattle HA template
# Setup nodes configuration. 3 nodes required for RKE HA.
#
nodes:
- address: <IP>
  role:
  - controlplane
  - etcd
  - worker
  user: <USER>
  ssh_key_path: <PEM_FILE>	# Optionally ssh_key: <PEM_KEY> could be used
#
# Optionally, any other RKE config object could be added here.
#
`
)

type TLSIssuerConfig struct {
	// Provider for managing TLS
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`
	// Issuer kind could be ca | acme
	Kind string `yaml:"kind,omitempty" json:"kind,omitempty"`
	// Pem tls ca key base64 value
	CAKey string `yaml:"ca_key,omitempty" json:"ca_key,omitempty"`
	// TLS ca key path
	CAKeyPath string `yaml:"ca_key_path,omitempty" json:"ca_key_path,omitempty"`
	// Pem tls cacert base64 value
	CACert string `yaml:"ca_cert,omitempty" json:"cert,omitempty"`
	// TLS cacert path
	CACertPath string `yaml:"ca_cert_path,omitempty" json:"ca_cert_path,omitempty"`
	// Acme email
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
	// Acme production environment?
	Production bool `yaml:"production,omitempty" json:"production,omitempty"`
	// Change default cert for ingress
	Default bool `yaml:"default,omitempty" json:"default,omitempty"`
}

func (config *TLSIssuerConfig) GetKind() string {
	return config.Kind
}

func (config *TLSIssuerConfig) GetProvider() string {
	return config.Provider
}

func (config *TLSIssuerConfig) IsDefault() bool {
	return config.Default
}

func (config *TLSIssuerConfig) SetDefault(b bool) {
	config.Default = false
}

func (config *TLSIssuerConfig) CheckCACert() error {
	var err error
	if config.CACertPath != "" || config.CACert != "" {
		// Update rancher CA cert base64 values from file if needed
		if config.CACertPath != "" {
			config.CACert, err = GetBase64FromFile(config.CACertPath)
			if err != nil {
				return err
			}
			config.CACertPath = ""
		}
		// Check rancher CA cert base64 string
		tempCert, err := CheckBase64String(config.CACert)
		if err != nil {
			return fmt.Errorf("CA cert incorrect Base64 string, %v", err)
		}
		// Check if rancher CA cert is valid CA
		_, err = verifyCert(tempCert)
		if err != nil {
			return err
		}
	}
	return nil
}

func (config *TLSIssuerConfig) CheckCAKey() error {
	var err error
	// Update CA Key base64 values from files if needed
	if config.CAKeyPath != "" {
		config.CAKey, err = GetBase64FromFile(config.CAKeyPath)
		if err != nil {
			return err
		}
		config.CAKeyPath = ""
	}
	// Check CA Key base64 string
	tempCAKey, err := CheckBase64String(config.CAKey)
	if err != nil {
		return fmt.Errorf("CA key incorrect Base64 string, %v", err)
	}
	// Verify ca Key PKCS1 format
	tempCAKey, err = verifyPkcs1(tempCAKey)
	if err != nil {
		return fmt.Errorf("CA key has incorrect format. PKCS1 and PKCS8 allowed, %v", err)
	}
	config.CAKey = GetBase64(tempCAKey)

	return nil
}

func (config *TLSIssuerConfig) Check() error {
	if config.Provider != templates.TLSManager {
		return fmt.Errorf("Supported TLS issuer providers %s", templates.TLSManager)
	}

	if config.Kind != "" && config.Kind != templates.TLSIssuerCA && config.Kind != templates.TLSIssuerAcme {
		return fmt.Errorf("TLS issuer kind should be `nil` | %s | %s", templates.TLSIssuerCA, templates.TLSIssuerAcme)
	}

	if config.Kind != templates.TLSIssuerCA {
		if config.IsDefault() {
			logrus.Warnf("Just TLS issuer kind %s could be set as default. Setting to false", templates.TLSIssuerCA)
			config.Default = false
		}
	}

	if config.Kind == templates.TLSIssuerCA {
		if (config.CAKeyPath == "" && config.CAKey == "") || (config.CACertPath == "" && config.CACert == "") {
			return fmt.Errorf("TLS issuer kind %s, requires TLS ca_key_path or ca_key and ca_path or ca", templates.TLSIssuerCA)
		}

		err := config.CheckCAKey()
		if err != nil {
			return err
		}
	}

	if config.Kind == templates.TLSIssuerAcme {
		// Check if Acme email is provided
		if config.Email == "" {
			return fmt.Errorf("TLS issuer kind %s require Email", templates.TLSIssuerAcme)
		}
		// Check if Acme production to load acme staging CA
		if !config.Production {
			bodyBytes, err := GetURL(AcmeStagingCA)
			if err != nil {
				return err
			}
			config.CACert = GetBase64(bodyBytes)
		}
	}

	// Checking CA cert
	err := config.CheckCACert()
	if err != nil {
		return err
	}

	return nil
}

type TLSConfig struct {
	// TLS mode for installing cattle HA
	Mode string `yaml:"mode,omitempty" json:"mode,omitempty"`
	// TLS issuer configuration
	Issuer *TLSIssuerConfig `yaml:"issuer,omitempty" json:"issuer,omitempty"`
	// Pem tls key base64 value used by cattle. Just for templates.TLSModeIngress
	Key string `yaml:"key,omitempty" json:"key,omitempty"`
	// TLS key path used by cattle. Just for templates.TLSModeIngress
	KeyPath string `yaml:"key_path,omitempty" json:"key_path,omitempty"`
	// Pem tls cert base64 value used by cattle. Just for templates.TLSModeIngress
	Cert string `yaml:"cert,omitempty" json:"cert,omitempty"`
	// TLS cert path used by cattle. Just for templates.TLSModeIngress
	CertPath string `yaml:"cert_path,omitempty" json:"cert_path,omitempty"`
}

func (config *TLSConfig) GetMode() string {
	return config.Mode
}

func (config *TLSConfig) Check(hostname string) error {
	if config.Mode != templates.TLSModeRancher && config.Mode != templates.TLSModeIngress && config.Mode != templates.TLSModeExternal {
		return fmt.Errorf("TLS mode should be %s | %s | %s ", templates.TLSModeRancher, templates.TLSModeIngress, templates.TLSModeExternal)
	}

	if config.Issuer.GetKind() != "" && config.Mode != templates.TLSModeIngress {
		return fmt.Errorf("TLS manager just valid for TLS mode %s", templates.TLSModeIngress)
	}

	if config.Mode == templates.TLSModeIngress {
		var err error
		if (config.Issuer.GetKind() == "") && ((config.KeyPath == "" && config.Key == "") || (config.CertPath == "" && config.Cert == "")) {
			return fmt.Errorf("TLS mode %s without manager requires TLS key_path or key and cert_path or cert", templates.TLSModeIngress)
		}

		if config.Issuer.GetKind() == "" {
			// Update rancher Key base64 values from files if needed
			if config.KeyPath != "" {
				config.Key, err = GetBase64FromFile(config.KeyPath)
				if err != nil {
					return err
				}
				config.KeyPath = ""
			}
			// Check rancher Key base64 string
			tempKey, err := CheckBase64String(config.Key)
			if err != nil {
				return fmt.Errorf("TLS key incorrect Base64 string, %v", err)
			}
			// Verify Key PKCS1 format
			tempKey, err = verifyPkcs1(tempKey)
			if err != nil {
				return fmt.Errorf("TLS key incorrect format PKCS1 and PKCS8 allowed, %v", err)
			}
			config.Key = GetBase64(tempKey)

			// Update rancher Cert base64 values from files if needed
			if config.CertPath != "" {
				config.Cert, err = GetBase64FromFile(config.CertPath)
				if err != nil {
					return err
				}
				config.CertPath = ""
			}
			// Check rancher Cert base64 string
			tempCert, err := CheckBase64String(config.Cert)
			if err != nil {
				return fmt.Errorf("TLS cert incorrect Base64 string, %v", err)
			}
			// Check rancher Cert is valid for hostname
			err = verifyCertHostname(tempCert, hostname)
			if err != nil {
				return err
			}
		}
	}

	return config.Issuer.Check()
}

type cattleConfig struct {
	// FQDN used for Cattle HA
	FQDN string `yaml:"fqdn" json:"fqdn"`
	// Docker server image
	Image string `yaml:"image" json:"image"`
	// Output mode
	Mode string `yaml:"mode" json:"mode"`
	// Namespace
	Namespace string `yaml:"namespace" json:"namespace"`
	// Replicas number
	Replicas int `yaml:"replicas" json:"replicas"`
	// TLS configuration for Cattle HA
	TLS *TLSConfig `yaml:"tls" json:"tls"`
}

func (config *cattleConfig) GetMode() string {
	return config.Mode
}

func (config *cattleConfig) Check() error {
	if config.FQDN == "" {
		return fmt.Errorf("No hostname defined")
	}

	if config.GetMode() != HAOutputRKE && config.Mode != HAOutputK8S {
		return fmt.Errorf("Output mode should be %s or %s", HAOutputRKE, HAOutputK8S)
	}

	if config.GetMode() == HAOutputK8S {
		if config.TLS.Mode == templates.TLSModeRancher {
			return fmt.Errorf("Output mode %s doesn't support TLS mode '%s' ", HAOutputK8S, templates.TLSModeRancher)
		}

		// If output mode HAOutputK8S issuer dafault should be false
		config.TLS.Issuer.SetDefault(false)
	}

	if config.Replicas < 1 {
		// Setting minimum replicas to 1
		config.Replicas = 1
	}

	return config.TLS.Check(config.FQDN)
}

func (config *cattleConfig) getExtras() []string {
	if config.TLS.Issuer.GetKind() != "" {
		return []string{CERTManagerURL}
	}
	return nil
}

func (config *cattleConfig) MarshalToString() (string, error) {
	result, err := yaml.Marshal(&config)
	if err != nil {
		logrus.Errorf("Marshaling cattle HA yaml, %v", err)
		return "", err
	}

	return string(result), nil
}

func (config *cattleConfig) GetManifest() (string, error) {
	return templates.CompileTemplateFromMap(templates.CattleHATemplate, config)
}

type cattleHAConfig struct {
	// RKE configuration for adding RKE config
	managementApi.RancherKubernetesEngineConfig `yaml:",inline" json:",inline"`
	// Cattle configuration for RKE addons
	Cattle *cattleConfig `yaml:"cattle" json:"cattle"`
}

func (config *cattleHAConfig) ReadFromFile(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		logrus.Errorf("Reading Rancher HA template %s, %v", filename, err)
		return err
	}

	err = yaml.Unmarshal(content, config)
	if err != nil {
		logrus.Errorf("Unmarshaling Rancher HA, %v", err)
		return err
	}

	if config.Cattle == nil {
		return fmt.Errorf("No Rancher HA info found at template %s", filename)
	}

	return nil
}

func (config *cattleHAConfig) Marshal() ([]byte, error) {
	tmpl := make(map[string]interface{})
	tmpl["cattle"] = config.Cattle
	result, err := yaml.Marshal(&tmpl)
	if err != nil {
		logrus.Errorf("Marshaling Rancher HA yaml, %v", err)
		return nil, err
	}

	if config.Cattle.GetMode() == HAOutputRKE {
		result = append([]byte(RKEHAMinimalTmpl), result...)
	}

	return result, nil
}

func (config *cattleHAConfig) MarshalToString() (string, error) {
	data, err := config.Marshal()

	return string(data), err
}

func (config *cattleHAConfig) MarshalToFile(filename string) error {
	result, err := config.Marshal()
	if err != nil {
		return err
	}

	err = WriteToFile([]byte(result), filename)
	if err != nil {
		return err
	}
	logrus.Infof("Saved Rancher HA %s template to %s", config.Cattle.Mode, filename)
	if config.Cattle.GetMode() == HAOutputRKE {
		logrus.Infof("To deploy %s cluster with Rancher HA, add 'nodes' info and run: \nrancher ha up -f %s", config.Cattle.GetMode(), filename)
	}
	return nil
}

func (config *cattleHAConfig) ManifestToFile(filename string) error {
	result, err := config.GetManifest()
	if err != nil {
		return err
	}

	err = WriteToFile([]byte(result), filename)
	if err != nil {
		return err
	}
	logrus.Infof("Saved Rancher HA %s deployment manifest to %s", config.Cattle.Mode, filename)
	if config.Cattle.GetMode() == HAOutputK8S {
		logrus.Infof("To deploy Rancher HA in your %s cluster run: \nkubectl create -f %s", config.Cattle.GetMode(), filename)
	}
	return nil
}

func (config *cattleHAConfig) GetExtras() (string, error) {
	extras := ""
	for _, extra := range config.Cattle.getExtras() {
		logrus.Infof("Getting extra url %s", extra)
		result, err := GetURL(extra)
		if err != nil {
			return "", err
		}
		extras = extras + string(result)
	}
	return extras, nil
}

func (config *cattleHAConfig) GetManifest() (string, error) {
	err := config.CheckManifest()
	if err != nil {
		logrus.Errorf("Checking Rancher HA manifest, %v", err)
		return "", err
	}
	manifest, err := config.Cattle.GetManifest()
	if err != nil {
		logrus.Errorf("Generating %s Rancher HA manifest, %v", config.Cattle.Mode, err)
		return "", err
	}

	extras, err := config.GetExtras()
	if err != nil {
		return "", err
	}
	manifest = manifest + "\n" + extras

	if config.Cattle.GetMode() == HAOutputRKE {
		// Getting RancherKubernetesEngineConfig output
		output := config.RancherKubernetesEngineConfig
		// Adding addons info
		output.Addons = output.Addons + manifest
		// Adding addons include for extras deployment
		//output.AddonsInclude = config.Cattle.getExtras()

		result, err := yaml.Marshal(&output)
		if err != nil {
			logrus.Errorf("Marshaling %s Rancher HA manifest, %v", config.Cattle.Mode, err)
			return "", err
		}
		manifest = string(result)
	}

	return manifest, nil
}

func (config *cattleHAConfig) CheckManifest() error {
	if config.Cattle.GetMode() == HAOutputRKE {
		// Configuring nginx-ingress-controller if TLSMode is passthrough
		if config.Cattle.TLS.GetMode() == templates.TLSModeRancher {
			if config.RancherKubernetesEngineConfig.Ingress.Provider == "" {
				config.RancherKubernetesEngineConfig.Ingress.Provider = "nginx"
			}
			if config.RancherKubernetesEngineConfig.Ingress.ExtraArgs == nil {
				config.RancherKubernetesEngineConfig.Ingress.ExtraArgs = map[string]string{}
			}
			config.RancherKubernetesEngineConfig.Ingress.ExtraArgs["enable-ssl-passthrough"] = ""
		}

		// Configuring nginx-ingress-controller if TLSIssuer is ca
		if (config.Cattle.TLS.GetMode() == templates.TLSModeIngress) && (config.Cattle.TLS.Issuer.GetKind() == templates.TLSIssuerCA) {
			if config.RancherKubernetesEngineConfig.Ingress.Provider == "" {
				config.RancherKubernetesEngineConfig.Ingress.Provider = "nginx"
			}
			if config.RancherKubernetesEngineConfig.Ingress.ExtraArgs == nil {
				config.RancherKubernetesEngineConfig.Ingress.ExtraArgs = map[string]string{}
			}
			if config.Cattle.TLS.Issuer.IsDefault() {
				config.RancherKubernetesEngineConfig.Ingress.ExtraArgs["default-ssl-certificate"] = config.Cattle.Namespace + "/" + templates.TLSIngressSecret
			}
		}

		// Checking at least one node is added
		nodeLen := len(config.RancherKubernetesEngineConfig.Nodes)
		if nodeLen == 0 {
			return fmt.Errorf("No nodes defined for %s cluster Rancher HA manifest. Recommended 3 for RKE HA", HAOutputRKE)
		}

		nodeRole := make(LabelCount)
		// Setting nodes SSHKey from SSHKeyPath to make cattleHAConfig self contained
		for node := range config.RancherKubernetesEngineConfig.Nodes {
			if config.RancherKubernetesEngineConfig.Nodes[node].SSHKeyPath == "" && config.RancherKubernetesEngineConfig.Nodes[node].SSHKey == "" {
				return fmt.Errorf("ssh_key_path or ssh_key should be defined. [%v]", config.RancherKubernetesEngineConfig.Nodes[node])
			}

			if config.RancherKubernetesEngineConfig.Nodes[node].SSHKeyPath != "" {
				key, err := GetStringFromFile(config.RancherKubernetesEngineConfig.Nodes[node].SSHKeyPath)
				if err != nil {
					return err
				}
				config.RancherKubernetesEngineConfig.Nodes[node].SSHKey = key
				config.RancherKubernetesEngineConfig.Nodes[node].SSHKeyPath = ""
			}

			// Counting nodes by RKE role
			for _, v := range config.RancherKubernetesEngineConfig.Nodes[node].Role {
				nodeRole.Increment(v)
			}
		}

		// Checking RKEMinimalEtcd etcd nodes are configured for RKE HA
		if nodeRole[RKEEtcdLabel] < RKEMinimalEtcd {
			logrus.Warnf("%d %s nodes defined. Minimum of %d %s required for RKE HA", nodeRole[RKEEtcdLabel], RKEEtcdLabel, RKEMinimalEtcd, RKEEtcdLabel)
		}

		for k, v := range nodeRole {
			logrus.Infof("Defined: %d %s nodes", v, k)
		}
	}

	return nil
}

func (config *cattleHAConfig) Check() error {
	if config.Cattle == nil {
		return fmt.Errorf("cattleHAConfig has no Cattle info")
	}

	err := config.Cattle.Check()
	if err != nil {
		return err
	}

	return nil
}

func HACommand() cli.Command {
	return cli.Command{
		Name:        "ha",
		Usage:       "Create Rancher HA minimal template and deployment files to rke and k8s",
		Description: HADescription,
		Subcommands: []cli.Command{
			{
				Name:        "template",
				Usage:       "Create Rancher HA minimal template",
				Description: HATemplateDesc,
				ArgsUsage:   "-f <rancher_HA_template> --hostname <FQDN> --tls-key <KEY_FILE> --tls-cert <CERT_FILE> --tls-ca <CA_FILE> ...",
				Action:      GetCattleHATemplate,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "hostname",
						Usage: "`FQDN` to access rancher. Required",
					},
					cli.IntFlag{
						Name:  "replicas",
						Usage: "Rancher deployment replicas `NUMBER`",
						Value: 1,
					},
					cli.StringFlag{
						Name:  "image",
						Usage: "Rancher docker image to deploy `IMAGE`",
						Value: RancherImage,
					},
					cli.StringFlag{
						Name:  "file,f",
						Usage: "Save Rancher HA template to a yaml `FILE`",
						Value: "rancher_HA.yml",
					},
					cli.StringFlag{
						Name:  "namespace",
						Usage: "Namespace to deploy Rancher HA `NAMESPACE`",
						Value: "cattle-system",
					},
					cli.StringFlag{
						Name:  "output-mode",
						Usage: "Output `MODE`, [ " + HAOutputRKE + " | " + HAOutputK8S + " ]. k8s mode generate deployment manifest.",
						Value: "rke",
					},
					cli.StringFlag{
						Name:  "output,o",
						Usage: "Save Rancher HA deployment `FILE`, just for " + HAOutputK8S + " output-mode. Default value '" + HAOutputK8S + "_" + templates.TLSModeIngress + "_file'",
					},
					cli.StringFlag{
						Name:  "tls-mode",
						Usage: "TLS `MODE`, [ " + templates.TLSModeIngress + " | " + templates.TLSModeRancher + " | " + templates.TLSModeExternal + " ]",
						Value: templates.TLSModeIngress,
					},
					cli.StringFlag{
						Name:  "tls-cert",
						Usage: "TLS cert `FILE` to use with rancher. Required for TLS mode " + templates.TLSModeIngress + "if TLS issuer none",
					},
					cli.StringFlag{
						Name:  "tls-key",
						Usage: "TLS key `FILE` to use with rancher. Required for TLS mode " + templates.TLSModeIngress + "if TLS issuer none",
					},
					cli.StringFlag{
						Name:  "tls-ca-cert",
						Usage: "TLS ca certs `FILE` to use with rancher. Required for TLS mode " + templates.TLSModeIngress + "if TLS issuer " + templates.TLSIssuerCA,
					},
					cli.StringFlag{
						Name:  "tls-ca-key",
						Usage: "TLS ca key `FILE` to use with tls-issuer " + templates.TLSIssuerCA + ". Required for TLS issuer " + templates.TLSIssuerCA,
					},
					cli.StringFlag{
						Name:  "tls-provider",
						Usage: "TLS cert `PROVIDER`, [ cert-manager ]. Valid on TLS mode " + templates.TLSModeIngress,
						Value: templates.TLSManager,
					},
					cli.StringFlag{
						Name:  "tls-issuer",
						Usage: "TLS cert `ISSUER`, [ " + templates.TLSIssuerCA + " | " + templates.TLSIssuerAcme + " ] managed by tls-manager. Valid on TLS mode " + templates.TLSModeIngress,
					},
					cli.BoolFlag{
						Name:  "tls-issuer-default",
						Usage: "Set tls-issuer as default cert issuer. Change ingress default cert on output-mode " + HAOutputRKE + ". Valid on TLS mode " + templates.TLSModeIngress,
					},
					cli.StringFlag{
						Name:  "tls-acme-email",
						Usage: "Email address `EMAIL` to use with tls-issuer " + templates.TLSIssuerAcme + ". Required for TLS issuer " + templates.TLSIssuerAcme,
					},
					cli.BoolFlag{
						Name:  "tls-acme-prod",
						Usage: "tls-issuer " + templates.TLSIssuerAcme + " production enviroment. Staging if false. Used for TLS issuer " + templates.TLSIssuerAcme,
					},
				},
			},
			{
				Name:      "up",
				Usage:     "Up new RKE cluster from Rancher HA template",
				ArgsUsage: "-f <cattle_HA_template> ...",
				Action:    HAUp,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "file,f",
						Usage: "Read Rancher HA minimal template from `FILE`",
						Value: "rancher_HA.yml",
					},
					cli.StringFlag{
						Name:  "output,o",
						Usage: "Save Rancher HA deployment `FILE`. Default value 'output-mode_tls-mode_file'",
					},
					cli.BoolFlag{
						Name:  "preview,p",
						Usage: "Preview mode. Just generate Rancher HA template don't up cluster",
					},
				},
			},
		},
	}
}

func newCattleHAConfig(ctx *cli.Context) *cattleHAConfig {
	return &cattleHAConfig{
		Cattle: &cattleConfig{
			FQDN:      ctx.String("hostname"),
			Image:     ctx.String("image"),
			Mode:      ctx.String("output-mode"),
			Namespace: ctx.String("namespace"),
			Replicas:  ctx.Int("replicas"),
			TLS: &TLSConfig{
				Mode:     ctx.String("tls-mode"),
				KeyPath:  ctx.String("tls-key"),
				CertPath: ctx.String("tls-cert"),
				Issuer: &TLSIssuerConfig{
					Provider:   ctx.String("tls-provider"),
					Kind:       ctx.String("tls-issuer"),
					CAKeyPath:  ctx.String("tls-ca-key"),
					CACertPath: ctx.String("tls-ca-cert"),
					Email:      ctx.String("tls-acme-email"),
					Production: ctx.Bool("tls-acme-prod"),
					Default:    ctx.Bool("tls-issuer-default"),
				},
			},
		},
	}
}

func GetCattleHATemplate(ctx *cli.Context) error {
	output := newCattleHAConfig(ctx)

	err := output.Check()
	if err != nil {
		return err
	}

	fileName := ctx.String("file")
	err = output.MarshalToFile(fileName)
	if err != nil {
		return err
	}

	if output.Cattle.Mode == HAOutputK8S {
		outFile := ctx.String("output")
		if len(outFile) == 0 {
			outFile = output.Cattle.Mode + "_" + output.Cattle.TLS.Mode + "_" + fileName
		}
		err = output.ManifestToFile(outFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func HAUp(ctx *cli.Context) error {
	fileName := ctx.String("file")
	if len(fileName) == 0 {
		return fmt.Errorf("Rancher HA minimal template file must be provided")
	}

	tmpl := &cattleHAConfig{}
	err := tmpl.ReadFromFile(fileName)
	if err != nil {
		return err
	}

	err = tmpl.Check()
	if err != nil {
		return err
	}

	outFile := ctx.String("output")
	if len(outFile) == 0 {
		outFile = tmpl.Cattle.Mode + "_" + tmpl.Cattle.TLS.Mode + "_" + fileName
	}

	err = tmpl.ManifestToFile(outFile)
	if err != nil {
		return err
	}

	if ctx.Bool("preview") || tmpl.Cattle.Mode != HAOutputRKE {
		return nil
	}

	arguments := []string{"up", "--config", outFile}
	err = RKERun(arguments)
	if err != nil {
		return err
	}

	logrus.Infof("RKE cluster deployed successfully\nRKE cluster config file - %s\nK8S config file - kube_config_%s\nRancher URL https://%s", outFile, outFile, tmpl.Cattle.FQDN)
	return nil
}

func CheckBase64String(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

func GetBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func GetBase64FromFile(file string) (string, error) {
	if file == "" {
		return "", fmt.Errorf("File name is empty")
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("Reading file %s, %v", file, err)
	}

	return GetBase64([]byte(data)), nil
}

func GetStringFromFile(file string) (string, error) {
	if file == "" {
		return "", fmt.Errorf("File name is empty")
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("Reading file %s, %v", file, err)
	}

	return string(data), nil
}

func GetURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func WriteToFile(content []byte, filename string) error {
	if len(filename) == 0 {
		logrus.Info("No filename provided writing to stdout")
		fmt.Printf("%s", content)
		return nil
	}

	err := ioutil.WriteFile(filename, content, 0644)
	if err != nil {
		logrus.Errorf("Writing filename %s, %v", filename, err)
		return err
	}

	return nil
}
