package cmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"
	"unicode"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/rancher/cli/cliclient"
	"github.com/rancher/cli/config"
	"github.com/rancher/norman/clientbase"
	ntypes "github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	letters             = "abcdefghijklmnopqrstuvwxyz0123456789"
	cfgFile             = "cli2.json"
	kubeConfigKeyFormat = "%s-%s"
)

var (
	errNoURL    = errors.New("RANCHER_URL environment or --Url is not set, run `login`")
	errNotFound = errors.New("not found")
	// ManagementResourceTypes lists the types we use the management client for
	ManagementResourceTypes = []string{"cluster", "node", "project"}
	// ProjectResourceTypes lists the types we use the cluster client for
	ProjectResourceTypes = []string{"secret", "namespacedSecret", "workload"}
	// ClusterResourceTypes lists the types we use the project client for
	ClusterResourceTypes = []string{"persistentVolume", "storageClass", "namespace"}

	formatFlag = cli.StringFlag{
		Name:  "format,o",
		Usage: "'json', 'yaml' or custom format",
	}

	quietFlag = cli.BoolFlag{
		Name:  "quiet,q",
		Usage: "Only display IDs or suppress help text",
	}
)

type MemberData struct {
	Name       string
	MemberType string
	AccessType string
}

type RoleTemplate struct {
	ID          string
	Name        string
	Description string
}

type RoleTemplateBinding struct {
	ID      string
	User    string
	Role    string
	Created string
}

func listAllRoles() []string {
	roles := []string{}
	roles = append(roles, ManagementResourceTypes...)
	roles = append(roles, ProjectResourceTypes...)
	roles = append(roles, ClusterResourceTypes...)
	return roles
}

func listRoles(ctx *cli.Context, context string) error {
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	filter := defaultListOpts(ctx)
	filter.Filters["hidden"] = false
	filter.Filters["context"] = context

	templates, err := c.ManagementClient.RoleTemplate.List(filter)
	if err != nil {
		return err
	}

	writer := NewTableWriter([][]string{
		{"ID", "ID"},
		{"NAME", "Name"},
		{"DESCRIPTION", "Description"},
	}, ctx)

	defer writer.Close()

	for _, item := range templates.Data {
		writer.Write(&RoleTemplate{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		})
	}

	return writer.Err()
}

func listRoleTemplateBindings(ctx *cli.Context, b []RoleTemplateBinding) error {
	writer := NewTableWriter([][]string{
		{"BINDING-ID", "ID"},
		{"USER", "User"},
		{"ROLE", "Role"},
		{"CREATED", "Created"},
	}, ctx)

	defer writer.Close()

	for _, item := range b {
		writer.Write(&RoleTemplateBinding{
			ID:      item.ID,
			User:    item.User,
			Role:    item.Role,
			Created: item.Created,
		})
	}
	return writer.Err()
}

func getKubeConfigForUser(ctx *cli.Context, user string) (*api.Config, error) {
	cf, err := loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	focusedServer := cf.FocusedServer()
	kubeConfig, _ := focusedServer.KubeConfigs[fmt.Sprintf(kubeConfigKeyFormat, user, focusedServer.FocusedCluster())]
	return kubeConfig, nil
}

func setKubeConfigForUser(ctx *cli.Context, user string, kubeConfig *api.Config) error {
	cf, err := loadConfig(ctx)
	if err != nil {
		return err
	}

	if cf.FocusedServer().KubeConfigs == nil {
		cf.FocusedServer().KubeConfigs = make(map[string]*api.Config)
	}
	focusedServer := cf.FocusedServer()
	focusedServer.KubeConfigs[fmt.Sprintf(kubeConfigKeyFormat, user, focusedServer.FocusedCluster())] = kubeConfig
	return cf.Write()
}

func usersToNameMapping(u []managementClient.User) map[string]string {
	userMapping := make(map[string]string)
	for _, user := range u {
		if user.Name != "" {
			userMapping[user.ID] = user.Name

		} else {
			userMapping[user.ID] = user.Username
		}
	}
	return userMapping
}

func searchForMember(ctx *cli.Context, c *cliclient.MasterClient, name string) (*managementClient.Principal, error) {
	filter := defaultListOpts(ctx)
	filter.Filters["ID"] = "thisisnotathingIhope"

	// A collection is needed to get the action link
	pCollection, err := c.ManagementClient.Principal.List(filter)
	if err != nil {
		return nil, err
	}

	p := managementClient.SearchPrincipalsInput{
		Name: name,
	}

	results, err := c.ManagementClient.Principal.CollectionActionSearch(pCollection, &p)
	if err != nil {
		return nil, err
	}

	dataLength := len(results.Data)
	switch {
	case dataLength == 0:
		return nil, fmt.Errorf("no results found for %q", name)
	case dataLength == 1:
		return &results.Data[0], nil
	case dataLength >= 10:
		results.Data = results.Data[:10]
	}

	var names []string

	for _, person := range results.Data {
		names = append(names, person.Name+fmt.Sprintf(" (%s)", person.PrincipalType))
	}
	selection := selectFromList("Multiple results found:", names)

	return &results.Data[selection], nil
}

func getRancherServerVersion(c *cliclient.MasterClient) (string, error) {
	setting, err := c.ManagementClient.Setting.ByID("server-version")
	if err != nil {
		return "", err
	}
	return setting.Value, err
}

func loadAndVerifyCert(path string) (string, error) {
	caCert, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return verifyCert(caCert)
}

func verifyCert(caCert []byte) (string, error) {
	// replace the escaped version of the line break
	caCert = bytes.Replace(caCert, []byte(`\n`), []byte("\n"), -1)

	block, _ := pem.Decode(caCert)

	if nil == block {
		return "", errors.New("No cert was found")
	}

	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	if !parsedCert.IsCA {
		return "", errors.New("CACerts is not valid")
	}
	return string(caCert), nil
}

func loadConfig(ctx *cli.Context) (config.Config, error) {
	// path will always be set by the global flag default
	path := ctx.GlobalString("config")
	path = filepath.Join(path, cfgFile)

	cf := config.Config{
		Path:    path,
		Servers: make(map[string]*config.ServerConfig),
	}

	content, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return cf, nil
	} else if err != nil {
		return cf, err
	}

	err = json.Unmarshal(content, &cf)
	cf.Path = path

	return cf, err
}

func lookupConfig(ctx *cli.Context) (*config.ServerConfig, error) {
	cf, err := loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	cs := cf.FocusedServer()
	if cs == nil {
		return nil, errors.New("no configuration found, run `login`")
	}

	return cs, nil
}

func GetClient(ctx *cli.Context) (*cliclient.MasterClient, error) {
	cf, err := lookupConfig(ctx)
	if err != nil {
		return nil, err
	}

	mc, err := cliclient.NewMasterClient(cf)
	if err != nil {
		return nil, err
	}

	return mc, nil
}

// GetResourceType maps an incoming resource type to a valid one from the schema
func GetResourceType(c *cliclient.MasterClient, resource string) (string, error) {
	if c.ManagementClient != nil {
		for key := range c.ManagementClient.APIBaseClient.Types {
			if strings.ToLower(key) == strings.ToLower(resource) {
				return key, nil
			}
		}
	}
	if c.ProjectClient != nil {
		for key := range c.ProjectClient.APIBaseClient.Types {
			if strings.ToLower(key) == strings.ToLower(resource) {
				return key, nil
			}
		}
	}
	if c.ClusterClient != nil {
		for key := range c.ClusterClient.APIBaseClient.Types {
			if strings.ToLower(key) == strings.ToLower(resource) {
				return key, nil
			}
		}
	}
	if c.CAPIClient != nil {
		for key := range c.CAPIClient.APIBaseClient.Types {
			lowerKey := strings.ToLower(key)
			if strings.HasPrefix(lowerKey, "cluster.x-k8s.io") && lowerKey == strings.ToLower(resource) {
				return key, nil
			}
		}
	}
	return "", fmt.Errorf("unknown resource type: %s", resource)
}

// Lookup looks up a resource by ID first, and then by name if it's not found by ID.
func Lookup(c *cliclient.MasterClient, name string, types ...string) (*ntypes.Resource, error) {
	var byName *ntypes.Resource

	for _, schemaType := range types {
		schemaClient, err := lookupSchemaClient(c, schemaType)
		if err != nil {
			return nil, err
		}
		// Attempt to get the resource by ID
		resource, err := lookupByIDWithClient(schemaClient, name, schemaType)
		if err != nil && !errors.Is(err, errNotFound) {
			return nil, err
		}
		if resource != nil {
			return resource, nil
		}

		// Resource was not found assuming the ID, check if it's the name of a resource
		resource, err = lookupByNameWithClient(schemaClient, name, schemaType)
		if err != nil && !errors.Is(err, errNotFound) {
			return nil, err
		}

		// No matches for this schemaType, try the next one
		if resource == nil {
			continue
		}

		if byName != nil {
			return nil, fmt.Errorf("multiple resources named %s: %s:%s, %s:%s", name, resource.Type,
				resource.ID, byName.Type, byName.ID)
		}

		byName = resource

	}

	if byName == nil {
		return nil, fmt.Errorf("resource %s: %w", name, errNotFound)
	}

	return byName, nil
}

func lookupSchemaClient(c *cliclient.MasterClient, schemaType string) (clientbase.APIBaseClientInterface, error) {
	rt, err := GetResourceType(c, schemaType)
	if err != nil {
		logrus.Debugf("Error GetResourceType: %v", err)
		return nil, err
	}
	var schemaClient clientbase.APIBaseClientInterface
	// the schemaType dictates which client we need to use
	if c.CAPIClient != nil {
		if strings.HasPrefix(rt, "cluster.x-k8s.io") {
			schemaClient = c.CAPIClient
		}
	}
	if c.ManagementClient != nil {
		if _, ok := c.ManagementClient.APIBaseClient.Types[rt]; ok {
			schemaClient = c.ManagementClient
		}
	}
	if c.ProjectClient != nil {
		if _, ok := c.ProjectClient.APIBaseClient.Types[rt]; ok {
			schemaClient = c.ProjectClient
		}
	}
	if c.ClusterClient != nil {
		if _, ok := c.ClusterClient.APIBaseClient.Types[rt]; ok {
			schemaClient = c.ClusterClient
		}
	}
	return schemaClient, nil
}

// LookupByName looks up a resource by name
func LookupByName(c *cliclient.MasterClient, name, schemaType string) (*ntypes.Resource, error) {
	schemaClient, err := lookupSchemaClient(c, schemaType)
	if err != nil {
		return nil, err
	}
	return lookupByNameWithClient(schemaClient, name, schemaType)
}

func lookupByNameWithClient(schemaClient clientbase.APIBaseClientInterface, name, schemaType string) (*ntypes.Resource, error) {
	var collection ntypes.ResourceCollection

	listOpts := &ntypes.ListOpts{
		Filters: map[string]interface{}{
			"name":         name,
			"removed_null": 1,
		},
	}

	if err := schemaClient.List(schemaType, listOpts, &collection); err != nil {
		logrus.Debugf("Error schemaClient.List: %v", err)
		if clientbase.IsNotFound(err) {
			return nil, fmt.Errorf("resource of type %s with name %s: %w", schemaType, name, errNotFound)
		}
		// prevents 403 from falsely returning as an unhandled error
		if apiError, ok := err.(*clientbase.APIError); ok {
			if apiError.StatusCode == http.StatusForbidden {
				return nil, fmt.Errorf("resource of type %s with name %s: %w", schemaType, name, errNotFound)
			}
		}
		return nil, err
	}

	if len(collection.Data) > 1 {
		ids := []string{}
		for _, data := range collection.Data {
			ids = append(ids, data.ID)
		}
		return nil, fmt.Errorf("multiple resources of type %s found for name %s: %v", schemaType, name, ids)
	}
	if len(collection.Data) == 0 {
		return nil, fmt.Errorf("resource of type %s with name %s: %w", schemaType, name, errNotFound)
	}

	return &collection.Data[0], nil
}

// LookupByID looks up a resource by it's ID
func LookupByID(c *cliclient.MasterClient, id, schemaType string) (*ntypes.Resource, error) {
	schemaClient, err := lookupSchemaClient(c, schemaType)
	if err != nil {
		return nil, err
	}
	return lookupByIDWithClient(schemaClient, id, schemaType)
}

func lookupByIDWithClient(schemaClient clientbase.APIBaseClientInterface, id, schemaType string) (*ntypes.Resource, error) {
	var resource ntypes.Resource

	if err := schemaClient.ByID(schemaType, id, &resource); err != nil {
		logrus.Debugf("Error schemaClient.ByID: %v", err)
		if clientbase.IsNotFound(err) {
			return nil, fmt.Errorf("resource of type %s with ID %s: %w", schemaType, id, errNotFound)
		}
		// prevents 403 from falsely returning as an unhandled error
		if apiError, ok := err.(*clientbase.APIError); ok {
			if apiError.StatusCode == http.StatusForbidden {
				return nil, fmt.Errorf("resource of type %s with ID %s: %w", schemaType, id, errNotFound)
			}
		}
		return nil, err
	}
	return &resource, nil
}

func RandomName() string {
	return strings.Replace(namesgenerator.GetRandomName(0), "_", "-", -1)
}

// RandomLetters returns a string with random letters of length n
func RandomLetters(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func appendTabDelim(buf *bytes.Buffer, value string) {
	if buf.Len() == 0 {
		buf.WriteString(value)
	} else {
		buf.WriteString("\t")
		buf.WriteString(value)
	}
}

func SimpleFormat(values [][]string) (string, string) {
	headerBuffer := bytes.Buffer{}
	valueBuffer := bytes.Buffer{}
	for _, v := range values {
		appendTabDelim(&headerBuffer, v[0])
		if strings.Contains(v[1], "{{") {
			appendTabDelim(&valueBuffer, v[1])
		} else {
			appendTabDelim(&valueBuffer, "{{."+v[1]+"}}")
		}
	}

	headerBuffer.WriteString("\n")
	valueBuffer.WriteString("\n")

	return headerBuffer.String(), valueBuffer.String()
}

func defaultAction(fn func(ctx *cli.Context) error) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		if ctx.Bool("help") {
			cli.ShowAppHelp(ctx)
			return nil
		}
		return fn(ctx)
	}
}

func printTemplate(out io.Writer, templateContent string, obj interface{}) error {
	funcMap := map[string]interface{}{
		"endpoint": FormatEndpoint,
		"ips":      FormatIPAddresses,
		"json":     FormatJSON,
	}
	tmpl, err := template.New("").Funcs(funcMap).Parse(templateContent)
	if err != nil {
		return err
	}

	return tmpl.Execute(out, obj)
}

func selectFromList(header string, choices []string) int {
	if header != "" {
		fmt.Println(header)
	}

	reader := bufio.NewReader(os.Stdin)
	selected := -1
	for selected <= 0 || selected > len(choices) {
		for i, choice := range choices {
			fmt.Printf("[%d] %s\n", i+1, choice)
		}
		fmt.Print("Select: ")

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		num, err := strconv.Atoi(text)
		if err == nil {
			selected = num
		}
	}
	return selected - 1
}

func processExitCode(err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			os.Exit(status.ExitStatus())
		}
	}

	return err
}

func SplitOnColon(s string) []string {
	return strings.Split(s, ":")
}

func parseClusterAndProjectName(name string) (string, string, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return "", "", fmt.Errorf("Unable to extract clustername and projectname from [%s]", name)
}

func parseClusterAndProjectID(id string) (string, string, error) {
	// Validate id
	// Examples:
	// c-qmpbm:p-mm62v
	// c-qmpbm:project-mm62v
	// c-m-j2s7m6lq:p-mm62v
	// See https://github.com/rancher/rancher/issues/14400
	if match, _ := regexp.MatchString("((local)|(c-[[:alnum:]]{5})|(c-m-[[:alnum:]]{8})):(p|project)-[[:alnum:]]{5}", id); match {
		parts := SplitOnColon(id)
		return parts[0], parts[1], nil
	}
	return "", "", fmt.Errorf("Unable to extract clusterid and projectid from [%s]", id)
}

// Return a JSON blob of the file at path
func readFileReturnJSON(path string) ([]byte, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}
	// This is probably already JSON if true
	if hasPrefix(file, []byte("{")) {
		return file, nil
	}
	return yaml.YAMLToJSON(file)
}

// renameKeys renames the keys in a given map of arbitrary depth with a provided function for string keys.
func renameKeys(input map[string]interface{}, f func(string) string) {
	for k, v := range input {
		delete(input, k)
		newKey := f(k)
		input[newKey] = v
		if innerMap, ok := v.(map[string]interface{}); ok {
			renameKeys(innerMap, f)
		}
	}
}

// convertSnakeCaseKeysToCamelCase takes a map and recursively transforms all snake_case keys into camelCase keys.
func convertSnakeCaseKeysToCamelCase(input map[string]interface{}) {
	renameKeys(input, convert.ToJSONKey)
}

// Return true if the first non-whitespace bytes in buf is prefix.
func hasPrefix(buf []byte, prefix []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, prefix)
}

func settingsToMap(client *cliclient.MasterClient) (map[string]string, error) {
	configMap := make(map[string]string)

	settings, err := client.ManagementClient.Setting.List(baseListOpts())
	if err != nil {
		return nil, err
	}

	for _, setting := range settings.Data {
		configMap[setting.Name] = setting.Value
	}

	return configMap, nil
}

// getClusterNames maps cluster ID to name and defaults to ID if name is blank
func getClusterNames(ctx *cli.Context, c *cliclient.MasterClient) (map[string]string, error) {
	clusterNames := make(map[string]string)
	clusterCollection, err := c.ManagementClient.Cluster.List(defaultListOpts(ctx))
	if err != nil {
		return clusterNames, err
	}

	for _, cluster := range clusterCollection.Data {
		if cluster.Name == "" {
			clusterNames[cluster.ID] = cluster.ID
		} else {
			clusterNames[cluster.ID] = cluster.Name
		}
	}
	return clusterNames, nil
}

func getClusterName(cluster *managementClient.Cluster) string {
	if cluster.Name != "" {
		return cluster.Name
	}
	return cluster.ID
}

func findStringInArray(s string, a []string) bool {
	for _, val := range a {
		if s == val {
			return true
		}
	}
	return false
}

func createdTimetoHuman(t string) (string, error) {
	parsedTime, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return "", err
	}
	return parsedTime.Format("02 Jan 2006 15:04:05 MST"), nil
}

func outputMembers(ctx *cli.Context, c *cliclient.MasterClient, members []managementClient.Member) error {
	writer := NewTableWriter([][]string{
		{"NAME", "Name"},
		{"MEMBER_TYPE", "MemberType"},
		{"ACCESS_TYPE", "AccessType"},
	}, ctx)

	defer writer.Close()

	for _, m := range members {
		principalID := m.UserPrincipalID
		if m.UserPrincipalID == "" {
			principalID = m.GroupPrincipalID
		}
		principal, err := c.ManagementClient.Principal.ByID(url.PathEscape(principalID))
		if err != nil {
			return err
		}
		writer.Write(&MemberData{
			Name:       principal.Name,
			MemberType: strings.Title(fmt.Sprintf("%s %s", principal.Provider, principal.PrincipalType)),
			AccessType: m.AccessType,
		})
	}
	return writer.Err()
}

func addMembersByNames(ctx *cli.Context, c *cliclient.MasterClient, members []managementClient.Member, toAddMembers []string, accessType string) ([]managementClient.Member, error) {
	for _, name := range toAddMembers {
		member, err := searchForMember(ctx, c, name)
		if err != nil {
			return nil, err
		}

		toAddMember := managementClient.Member{
			AccessType: accessType,
		}
		if member.PrincipalType == "user" {
			toAddMember.UserPrincipalID = member.ID
		} else {
			toAddMember.GroupPrincipalID = member.ID
		}
		members = append(members, toAddMember)
	}
	return members, nil
}

func deleteMembersByNames(ctx *cli.Context, c *cliclient.MasterClient, members []managementClient.Member, todeleteMembers []string) ([]managementClient.Member, error) {
	for _, name := range todeleteMembers {
		member, err := searchForMember(ctx, c, name)
		if err != nil {
			return nil, err
		}

		var toKeepMembers []managementClient.Member
		for _, m := range members {
			if m.GroupPrincipalID != member.ID && m.UserPrincipalID != member.ID {
				toKeepMembers = append(toKeepMembers, m)
			}
		}
		members = toKeepMembers
	}
	return members, nil
}

func tickerContext(ctx context.Context, duration time.Duration) <-chan time.Time {
	ticker := time.NewTicker(duration)
	go func() {
		<-ctx.Done()
		ticker.Stop()
	}()
	return ticker.C
}
