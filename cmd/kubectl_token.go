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
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	url2 "net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/cli/config"
	apiv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli"
	"golang.org/x/term"
)

const deleteExample = `
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

var oauthProviders = map[string]bool{
	"azureADProvider": true,
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

	// oauth providers
	"azureADProvider": true,
}

func CredentialCommand() cli.Command {
	configDir, err := ConfigDir()
	if err != nil {
		if runtime.GOOS == "windows" {
			configDir = "%HOME%\\.rancher"
		} else {
			configDir = "${HOME}/.rancher"
		}
	}
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
			{
				Name:   "delete",
				Usage:  fmt.Sprintf("Delete cached token used for kubectl login at [%s] \n %s", configDir, deleteExample),
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
	cachedCred, err := loadCachedCredential(ctx, cachedCredName)
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

	if err := cacheCredential(ctx, newCred, fmt.Sprintf("%s_%s", userID, clusterID)); err != nil {
		customPrint(fmt.Errorf("CacheToken: %v", err))
	}

	return json.NewEncoder(os.Stdout).Encode(newCred)
}

func deleteCachedCredential(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	// dir is always set by global default.
	dir := ctx.GlobalString("config")

	if len(cf.Servers) == 0 {
		customPrint(fmt.Sprintf("there are no cached tokens in [%s]", dir))
		return nil
	}

	if ctx.Args().First() == "all" {
		customPrint(fmt.Sprintf("removing cached tokens in [%s]", dir))
		for _, server := range cf.Servers {
			server.KubeCredentials = make(map[string]*config.ExecCredential)
		}
		return cf.Write()
	}

	for _, key := range ctx.Args() {
		customPrint(fmt.Sprintf("removing [%s]", key))
		for _, server := range cf.Servers {
			server.KubeCredentials[key] = nil
		}
	}

	return cf.Write()
}

func loadCachedCredential(ctx *cli.Context, key string) (*config.ExecCredential, error) {
	sc, err := lookupServerConfig(ctx)
	if err != nil {
		return nil, err
	}

	cred := sc.KubeToken(key)
	if cred == nil {
		return cred, nil
	}
	ts := cred.Status.ExpirationTimestamp
	if ts != nil && ts.Time.Before(time.Now()) {
		cf, err := loadConfig(ctx)
		if err != nil {
			return nil, err
		}
		cf.Servers[ctx.String("server")].KubeCredentials[key] = nil
		if err := cf.Write(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	return cred, nil
}

// there is overlap between this and the lookupConfig() function. However, lookupConfig() requires
// a server to be previously set in the Config, which might not be the case if rancher token
// is run before rancher login. Perhaps we can depricate rancher token down the line and defer
// all it does to login.
func lookupServerConfig(ctx *cli.Context) (*config.ServerConfig, error) {
	server := ctx.String("server")
	if server == "" {
		return nil, fmt.Errorf("name of rancher server is required")
	}

	cf, err := loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	sc := cf.Servers[server]
	if sc == nil {
		sc = &config.ServerConfig{
			KubeCredentials: make(map[string]*config.ExecCredential),
		}
		cf.Servers[server] = sc
		if err := cf.Write(); err != nil {
			return nil, err
		}
	}
	return sc, nil
}

func cacheCredential(ctx *cli.Context, cred *config.ExecCredential, id string) error {
	// cache only if valid
	if cred.Status.Token == "" {
		return nil
	}

	server := ctx.String("server")
	if server == "" {
		return fmt.Errorf("name of rancher server is required")
	}

	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	sc, err := lookupServerConfig(ctx)
	if err != nil {
		return err
	}

	if sc.KubeCredentials[id] == nil {
		sc.KubeCredentials = make(map[string]*config.ExecCredential)
	}
	sc.KubeCredentials[id] = cred
	cf.Servers[server] = sc
	return cf.Write()
}

func loginAndGenerateCred(input *LoginInput) (*config.ExecCredential, error) {
	// setup a client with the provided TLS configuration
	client, err := getClient(input.skipVerify, input.caCerts)
	if err != nil {
		return nil, err
	}

	authProviders, err := getAuthProviders(input.server)
	if err != nil {
		return nil, err
	}

	selectedProvider, err := selectAuthProvider(authProviders, input.authProvider)
	if err != nil {
		return nil, err
	}
	input.authProvider = selectedProvider.GetType()

	token := managementClient.Token{}
	if samlProviders[input.authProvider] {
		token, err = samlAuth(input, client)
		if err != nil {
			return nil, err
		}
	} else if oauthProviders[input.authProvider] {
		tokenPtr, err := oauthAuth(input, selectedProvider)
		if err != nil {
			return nil, err
		}
		token = *tokenPtr
	} else {
		customPrint(fmt.Sprintf("Enter credentials for %s \n", input.authProvider))
		token, err = basicAuth(input)
		if err != nil {
			return nil, err
		}
	}

	cred := &config.ExecCredential{
		TypeMeta: config.TypeMeta{
			Kind:       "ExecCredential",
			APIVersion: "client.authentication.k8s.io/v1beta1",
		},
		Status: &config.ExecCredentialStatus{},
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
	cred.Status.ExpirationTimestamp = &config.Time{Time: ts}
	return cred, nil

}

func basicAuth(input *LoginInput) (managementClient.Token, error) {
	token := managementClient.Token{}
	username, err := customPrompt("Enter username: ", true)
	if err != nil {
		return token, err
	}

	password, err := customPrompt("Enter password: ", false)
	if err != nil {
		return token, err
	}

	responseType := "kubeconfig"
	if input.clusterID != "" {
		responseType = fmt.Sprintf("%s_%s", responseType, input.clusterID)
	}

	body := fmt.Sprintf(`{"responseType":%q, "username":%q, "password":%q}`, responseType, username, password)

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

func samlAuth(input *LoginInput, client *http.Client) (managementClient.Token, error) {
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

	client.Timeout = 300 * time.Second

	loginRequest := fmt.Sprintf("%s/dashboard/auth/login?requestId=%s&publicKey=%s&responseType=%s",
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
			content, err := io.ReadAll(res.Body)
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

			client.Timeout = 150 * time.Second

			res, err = client.Do(req)
			if err != nil {
				// log error and use the token if login succeeds
				customPrint(fmt.Errorf("DeleteToken: %v", err))
			}
			defer res.Body.Close()

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

type TypedProvider interface {
	GetType() string
}

func getAuthProviders(server string) ([]TypedProvider, error) {
	authProviders := fmt.Sprintf(authProviderURL, server)
	customPrint(authProviders)

	response, err := request(http.MethodGet, authProviders, nil)
	if err != nil {
		return nil, err
	}

	if !gjson.ValidBytes(response) {
		return nil, errors.New("invalid JSON input")
	}

	data := gjson.GetBytes(response, "data").Array()

	supportedProviders := []TypedProvider{}
	for _, provider := range data {
		providerType := provider.Get("type").String()

		if providerType != "" && supportedAuthProviders[providerType] {
			var typedProvider TypedProvider

			switch providerType {
			case "azureADProvider":
				typedProvider = &apiv3.AzureADProvider{}
			case "localProvider":
				typedProvider = &apiv3.LocalProvider{}
			default:
				typedProvider = &apiv3.AuthProvider{}
			}

			err = json.Unmarshal([]byte(provider.Raw), typedProvider)
			if err != nil {
				return nil, err
			}
			supportedProviders = append(supportedProviders, typedProvider)
		}
	}

	return supportedProviders, err
}

func selectAuthProvider(authProviders []TypedProvider, providerType string) (TypedProvider, error) {
	if len(authProviders) == 0 {
		return nil, fmt.Errorf("no auth provider configured")
	}

	// if providerType was specified, look for it
	if providerType != "" {
		for _, p := range authProviders {
			if p.GetType() == providerType {
				return p, nil
			}
		}
		return nil, fmt.Errorf("provider %s not found", providerType)
	}

	// otherwise ask to the user (if more than one)
	if len(authProviders) == 1 {
		return authProviders[0], nil
	}

	var providers []string
	for i, val := range authProviders {
		providers = append(providers, fmt.Sprintf("%d - %s", i, val.GetType()))
	}

	try := 0
	for try < 3 {
		customPrint(fmt.Sprintf("Auth providers:\n%v", strings.Join(providers, "\n")))
		providerIndexStr, err := customPrompt("Select auth provider: ", true)
		if err != nil {
			try++
			continue
		}

		providerIndex, err := strconv.Atoi(providerIndexStr)
		if err != nil || (providerIndex < 0 || providerIndex > len(providers)-1) {
			customPrint("Pick a valid auth provider")
			try++
			continue
		}

		return authProviders[providerIndex], nil
	}

	return nil, fmt.Errorf("invalid auth provider")
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

// getClient return a client with the provided TLS configuration
func getClient(skipVerify bool, caCerts string) (*http.Client, error) {
	tlsConfig, err := getTLSConfig(skipVerify, caCerts)
	if err != nil {
		return nil, err
	}

	// clone the DefaultTransport to get the default values
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	return &http.Client{Transport: transport}, nil
}

func getTLSConfig(skipVerify bool, caCerts string) (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: skipVerify,
	}

	if caCerts == "" {
		return config, nil
	}

	// load custom certs
	cert, err := loadAndVerifyCert(caCerts)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(cert))
	if !ok {
		return nil, err
	}
	config.RootCAs = roots

	return config, nil
}

func request(method, url string, body io.Reader) ([]byte, error) {
	var response []byte

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	client, err := getClient(true, "")
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	response, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func customPrompt(msg string, show bool) (result string, err error) {
	fmt.Fprint(os.Stderr, msg)
	if show {
		_, err = fmt.Fscan(os.Stdin, &result)
	} else {
		var data []byte
		data, err = term.ReadPassword(int(os.Stdin.Fd()))
		result = string(data)
		fmt.Fprintf(os.Stderr, "\n")
	}
	return result, err
}

func customPrint(data interface{}) {
	fmt.Fprintf(os.Stderr, "%v \n", data)
}
