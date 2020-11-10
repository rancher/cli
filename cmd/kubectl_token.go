package cmd

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	url2 "net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/rancher/norman/types/convert"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

const deleteCommandUsage = `
Delete cached token used for kubectl login at ${PWD}/.cache/token

Example:
	# Delete a cached credential 
	$ rancher token delete cluster1_c-1234

	# Delete multiple cached credentials
	$ rancher token delete cluster1_c-1234 cluster2_c-2345

	# Delete all credentials 
	$ rancher token delete all
`

type LoginInput struct {
	server       string
	userID       string
	clusterID    string
	authProvider string
	caCerts      string
	skipVerify   bool
}

const (
	kubeConfigCache = "/.cache/token"
	cachedFileExt   = ".json"
	authProviderURL = "%s/v3-public/authProviders"
	authTokenURL    = "%s/v3-public/authTokens/%s"
)

var samlProviders = map[string]bool{
	"pingProvider":       true,
	"adfsProvider":       true,
	"keyCloakProvider":   true,
	"oktaProvider":       true,
	"shibbolethProvider": true,
}

var supportedAuthProviders = map[string]bool{
	"localProvider":           true,
	"freeIpaProvider":         true,
	"openLdapProvider":        true,
	"activeDirectoryProvider": true,

	// all saml providers
	"pingProvider":       true,
	"adfsProvider":       true,
	"keyCloakProvider":   true,
	"oktaProvider":       true,
	"shibbolethProvider": true,
}

func CredentialCommand() cli.Command {
	return cli.Command{
		Name:   "token",
		Usage:  "Authenticate and generate new kubeconfig token",
		Action: runCredential,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "server",
				Usage: "Name of rancher server",
			},
			cli.StringFlag{
				Name:  "user",
				Usage: "user-id",
			},
			cli.StringFlag{
				Name:  "cluster",
				Usage: "cluster-id",
			},
			cli.StringFlag{
				Name:  "auth-provider",
				Usage: "Name of Auth Provider to use for authentication",
			},
			cli.StringFlag{
				Name:  "cacerts",
				Usage: "Location of CaCerts to use",
			},
			cli.BoolFlag{
				Name:  "skip-verify",
				Usage: "Skip verification of the CACerts presented by the Server",
			},
		},
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "delete",
				Usage:  deleteCommandUsage,
				Action: deleteCachedCredential,
			},
		},
	}
}

func runCredential(ctx *cli.Context) error {
	if ctx.Bool("delete") {
		return deleteCachedCredential(ctx)
	}
	server := ctx.String("server")
	if server == "" {
		return fmt.Errorf("name of rancher server is required")
	}
	url, err := url2.Parse(server)
	if err != nil {
		return err
	}
	if url.Scheme == "" {
		server = fmt.Sprintf("https://%s", server)
	}
	userID := ctx.String("user")
	if userID == "" {
		return fmt.Errorf("user-id is required")
	}
	clusterID := ctx.String("cluster")

	cachedCredName := fmt.Sprintf("%s_%s", userID, clusterID)
	cachedCred, err := loadCachedCredential(cachedCredName)
	if err != nil {
		customPrint(fmt.Errorf("LoadToken: %v", err))
	}
	if cachedCred != nil {
		return json.NewEncoder(os.Stdout).Encode(cachedCred)
	}

	input := &LoginInput{
		server:       server,
		userID:       userID,
		clusterID:    clusterID,
		authProvider: ctx.String("auth-provider"),
		caCerts:      ctx.String("cacerts"),
		skipVerify:   ctx.Bool("skip-verify"),
	}

	newCred, err := loginAndGenerateCred(input)
	if err != nil {
		return err
	}

	if err := cacheCredential(newCred, fmt.Sprintf("%s_%s", userID, clusterID)); err != nil {
		customPrint(fmt.Errorf("CacheToken: %v", err))
	}

	return json.NewEncoder(os.Stdout).Encode(newCred)
}

func deleteCachedCredential(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	cacheDir := filepath.Join(dir, kubeConfigCache)
	if ctx.Args().First() == "all" {
		customPrint(fmt.Sprintf("removing cached tokens [%s]", cacheDir))
		return os.RemoveAll(cacheDir)
	}
	for _, key := range ctx.Args() {
		cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s%s", key, cachedFileExt))
		customPrint(fmt.Sprintf("removing [%s]", cachePath))
		err := os.Remove(cachePath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func loadCachedCredential(key string) (*ExecCredential, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	cachePath := filepath.Join(dir, kubeConfigCache, fmt.Sprintf("%s%s", key, cachedFileExt))
	f, err := os.Open(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	defer f.Close()
	var execCredential *ExecCredential
	if err := json.NewDecoder(f).Decode(&execCredential); err != nil {
		return nil, err
	}
	ts := execCredential.Status.ExpirationTimestamp
	if ts != nil && ts.Time.Before(time.Now()) {
		err = os.Remove(cachePath)
		return nil, err
	}
	return execCredential, nil
}

func cacheCredential(cred *ExecCredential, id string) error {
	// cache only if valid
	if cred.Status.Token == "" {
		return nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	cachePathDir := filepath.Join(dir, kubeConfigCache)
	if err := os.MkdirAll(cachePathDir, os.FileMode(0700)); err != nil {
		return err
	}
	path := filepath.Join(cachePathDir, fmt.Sprintf("%s%s", id, cachedFileExt))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(0600))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(cred)
}

func loginAndGenerateCred(input *LoginInput) (*ExecCredential, error) {
	if input.authProvider == "" {
		provider, err := getAuthProvider(input.server)
		if err != nil {
			return nil, err
		}
		input.authProvider = provider
	}
	tlsConfig, err := getTLSConfig(input)
	if err != nil {
		return nil, err
	}
	token := managementClient.Token{}
	if samlProviders[input.authProvider] {
		token, err = samlAuth(input, tlsConfig)
		if err != nil {
			return nil, err
		}
	} else {
		customPrint(fmt.Sprintf("Enter credentials for %s \n", input.authProvider))
		token, err = basicAuth(input, tlsConfig)
		if err != nil {
			return nil, err
		}
	}
	cred := &ExecCredential{
		TypeMeta: TypeMeta{
			Kind:       "ExecCredential",
			APIVersion: "client.authentication.k8s.io/v1beta1",
		},
		Status: &ExecCredentialStatus{},
	}
	cred.Status.Token = token.Token
	if token.ExpiresAt == "" {
		return cred, nil
	}
	ts, err := time.Parse(time.RFC3339, token.ExpiresAt)
	if err != nil {
		customPrint(fmt.Sprintf("\n error parsing time %s %v", token.ExpiresAt, err))
		return nil, err
	}
	cred.Status.ExpirationTimestamp = &Time{Time: ts}
	return cred, nil

}

func basicAuth(input *LoginInput, tlsConfig *tls.Config) (managementClient.Token, error) {
	token := managementClient.Token{}
	username, err := customPrompt("username", true)
	if err != nil {
		return token, err
	}

	password, err := customPrompt("password", false)
	if err != nil {
		return token, err
	}

	responseType := "kubeconfig"
	if input.clusterID != "" {
		responseType = fmt.Sprintf("%s_%s", responseType, input.clusterID)
	}

	body := fmt.Sprintf(`{"responseType":"%s", "username":"%s", "password":"%s"}`, responseType, username, password)

	url := fmt.Sprintf("%s/v3-public/%ss/%s?action=login", input.server, input.authProvider,
		strings.ToLower(strings.Replace(input.authProvider, "Provider", "", 1)))

	response, err := request(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return token, nil
	}

	apiError := map[string]interface{}{}
	err = json.Unmarshal(response, &apiError)
	if err != nil {
		return token, err
	}

	if responseType := apiError["type"]; responseType == "error" {
		return token, fmt.Errorf("error logging in: code: "+
			"[%v] message:[%v]", apiError["code"], apiError["message"])
	}

	err = json.Unmarshal(response, &token)
	if err != nil {
		return token, err
	}
	return token, nil
}

func samlAuth(input *LoginInput, tlsConfig *tls.Config) (managementClient.Token, error) {
	token := managementClient.Token{}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return token, err
	}
	publicKey := privateKey.PublicKey
	marshalKey, err := json.Marshal(publicKey)
	if err != nil {
		return token, err
	}
	encodedKey := base64.StdEncoding.EncodeToString(marshalKey)

	id, err := generateKey()
	if err != nil {
		return token, err
	}

	responseType := "kubeconfig"
	if input.clusterID != "" {
		responseType = fmt.Sprintf("%s_%s", responseType, input.clusterID)
	}

	tokenURL := fmt.Sprintf(authTokenURL, input.server, id)

	req, err := http.NewRequest(http.MethodGet, tokenURL, bytes.NewBuffer(nil))
	if err != nil {
		return token, err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{Transport: tr, Timeout: 300 * time.Second}

	loginRequest := fmt.Sprintf("%s/login?requestId=%s&publicKey=%s&responseType=%s",
		input.server, id, encodedKey, responseType)

	customPrint(fmt.Sprintf("\nLogin to Rancher Server at %s \n", loginRequest))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// timeout for user to login and get token
	timeout := time.NewTicker(15 * time.Minute)
	defer timeout.Stop()

	poll := time.NewTicker(10 * time.Second)
	defer poll.Stop()

	for {
		select {
		case <-poll.C:
			res, err := client.Do(req)
			if err != nil {
				return token, err
			}
			content, err := ioutil.ReadAll(res.Body)
			if err != nil {
				res.Body.Close()
				return token, err
			}
			res.Body.Close()
			err = json.Unmarshal(content, &token)
			if err != nil {
				return token, err
			}
			if token.Token == "" {
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(token.Token)
			if err != nil {
				return token, err
			}
			decryptedBytes, err := privateKey.Decrypt(nil, decoded, &rsa.OAEPOptions{Hash: crypto.SHA256})
			if err != nil {
				panic(err)
			}
			token.Token = string(decryptedBytes)

			// delete token
			req, err = http.NewRequest(http.MethodDelete, tokenURL, bytes.NewBuffer(nil))
			if err != nil {
				return token, err
			}
			req.Header.Set("content-type", "application/json")
			req.Header.Set("accept", "application/json")
			tr := &http.Transport{
				TLSClientConfig: tlsConfig,
			}
			client = &http.Client{Transport: tr, Timeout: 150 * time.Second}
			res, err = client.Do(req)
			if err != nil {
				// log error and use the token if login succeeds
				customPrint(fmt.Errorf("DeleteToken: %v", err))
			}
			return token, nil

		case <-timeout.C:
			break

		case <-interrupt:
			customPrint("received interrupt")
			break
		}

		return token, nil
	}
}

func getAuthProviders(server string) (map[string]string, error) {
	authProviders := fmt.Sprintf(authProviderURL, server)
	customPrint(authProviders)
	response, err := request(http.MethodGet, authProviders, nil)
	data := map[string]interface{}{}
	err = json.Unmarshal(response, &data)
	if err != nil {
		return nil, err
	}
	providers := map[string]string{}
	i := 0
	for _, value := range convert.ToMapSlice(data["data"]) {
		provider := convert.ToString(value["type"])
		if provider != "" && supportedAuthProviders[provider] {
			providers[fmt.Sprintf("%v", i)] = provider
			i++
		}
	}
	return providers, err
}

func getAuthProvider(server string) (string, error) {
	authProviders, err := getAuthProviders(server)
	if err != nil || authProviders == nil {
		return "", err
	}
	if len(authProviders) == 0 {
		return "", fmt.Errorf("no auth provider configured")
	}
	if len(authProviders) == 1 {
		return authProviders["0"], nil
	}
	try := 0
	var providers []string
	for key, val := range authProviders {
		providers = append(providers, fmt.Sprintf("%s - %s", key, val))
	}
	for try < 3 {
		provider, err := customPrompt(fmt.Sprintf("auth provider\n%v",
			strings.Join(providers, "\n")), true)
		if err != nil {
			try++
			continue
		}
		if _, ok := authProviders[provider]; !ok {
			customPrint("pick valid auth provider")
			try++
			continue
		}
		provider = authProviders[provider]
		return provider, nil
	}

	return "", fmt.Errorf("invalid auth provider")

}

func generateKey() (string, error) {
	characters := "abcdfghjklmnpqrstvwxz12456789"
	tokenLength := 32
	token := make([]byte, tokenLength)
	for i := range token {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			return "", err
		}
		token[i] = characters[r.Int64()]
	}

	return string(token), nil
}

func getTLSConfig(input *LoginInput) (*tls.Config, error) {
	config := &tls.Config{}
	if input.skipVerify || input.caCerts == "" {
		config = &tls.Config{
			InsecureSkipVerify: true,
		}
		return config, nil
	}

	if input.caCerts != "" {
		cert, err := loadAndVerifyCert(input.caCerts)
		if err != nil {
			return nil, err
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(cert))
		if !ok {
			return nil, err
		}
		config.RootCAs = roots
	}

	return config, nil
}

func request(method, url string, body io.Reader) ([]byte, error) {
	var response []byte
	var client *http.Client
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return response, err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	response, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func customPrompt(field string, show bool) (result string, err error) {
	fmt.Fprintf(os.Stderr, "Enter %s: ", field)
	if show {
		_, err = fmt.Fscan(os.Stdin, &result)
	} else {
		var data []byte
		data, err = terminal.ReadPassword(int(os.Stdin.Fd()))
		result = string(data)
		fmt.Fprintf(os.Stderr, "\n")
	}
	return result, err

}

func customPrint(data interface{}) {
	fmt.Fprintf(os.Stderr, "%v \n", data)
}

// ExecCredential is used by exec-based plugins to communicate credentials to
// HTTP transports. //v1beta1/types.go
type ExecCredential struct {
	TypeMeta `json:",inline"`

	// Spec holds information passed to the plugin by the transport. This contains
	// request and runtime specific information, such as if the session is interactive.
	Spec ExecCredentialSpec `json:"spec,omitempty"`

	// Status is filled in by the plugin and holds the credentials that the transport
	// should use to contact the API.
	// +optional
	Status *ExecCredentialStatus `json:"status,omitempty"`
}

// ExecCredentialSpec holds request and runtime specific information provided by
// the transport.
type ExecCredentialSpec struct{}

// ExecCredentialStatus holds credentials for the transport to use.
// Token and ClientKeyData are sensitive fields. This data should only be
// transmitted in-memory between client and exec plugin process. Exec plugin
// itself should at least be protected via file permissions.
type ExecCredentialStatus struct {
	// ExpirationTimestamp indicates a time when the provided credentials expire.
	// +optional
	ExpirationTimestamp *Time `json:"expirationTimestamp,omitempty"`
	// Token is a bearer token used by the client for request authentication.
	Token string `json:"token,omitempty"`
	// PEM-encoded client TLS certificates (including intermediates, if any).
	ClientCertificateData string `json:"clientCertificateData,omitempty"`
	// PEM-encoded private key for the above certificate.
	ClientKeyData string `json:"clientKeyData,omitempty"`
}

// TypeMeta describes an individual object in an API response or request
// with strings representing the type of the object and its API schema version.
// Structures that are versioned or persisted should inline TypeMeta.
type TypeMeta struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
}

// Time is a wrapper around time.Time which supports correct
// marshaling to YAML and JSON.  Wrappers are provided for many
// of the factory methods that the time package offers.
type Time struct {
	time.Time `protobuf:"-"`
}
