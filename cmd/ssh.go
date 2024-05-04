package cmd

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/urfave/cli"
)

const sshDescription = `
For any nodes created through Rancher using docker-machine,
you can SSH into the node. This is not supported for any custom nodes.
Examples:
	# SSH into a node by ID/name
	$ rancher ssh nodeFoo
	# SSH into a node by ID/name using the external IP address
	$ rancher ssh -e nodeFoo
	# SSH into a node by name but specify the login name to use
	$ rancher ssh -l login1 nodeFoo
	# SSH into a node by specifying login name and node using the @ syntax while adding a command to run
	$ rancher ssh login1@nodeFoo -- netstat -p tcp
`

func SSHCommand() cli.Command {
	return cli.Command{
		Name:        "ssh",
		Usage:       "SSH into a node",
		Description: sshDescription,
		Action:      nodeSSH,
		ArgsUsage:   "[NODE_ID/NODE_NAME]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "external,e",
				Usage: "Use the external ip address of the node",
			},
			cli.StringFlag{
				Name:  "login,l",
				Usage: "The login name",
			},
		},
	}
}

func nodeSSH(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "ssh")
	}

	if ctx.NArg() == 0 {
		return cli.ShowCommandHelp(ctx, "ssh")
	}

	user := ctx.String("login")
	nodeName := ctx.Args().First()

	if strings.Contains(nodeName, "@") {
		user = strings.Split(nodeName, "@")[0]
		nodeName = strings.Split(nodeName, "@")[1]
	}

	args = args[1:]

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	sshNode, key, err := getNodeAndKey(ctx, c, nodeName)
	if err != nil {
		return err
	}

	if user == "" {
		user = sshNode.SshUser
	}
	ipAddress := sshNode.IPAddress
	if ctx.Bool("external") {
		ipAddress = sshNode.ExternalIPAddress
	}

	return processExitCode(callSSH(key, ipAddress, user, args))
}

func getNodeAndKey(ctx *cli.Context, c *cliclient.MasterClient, nodeName string) (managementClient.Node, []byte, error) {
	sshNode := managementClient.Node{}
	resource, err := Lookup(c, nodeName, "node")
	if err != nil {
		return sshNode, nil, err
	}

	sshNode, err = getNodeByID(ctx, c, resource.ID)
	if err != nil {
		return sshNode, nil, err
	}

	link := sshNode.Links["nodeConfig"]
	if link == "" {
		// Get the machine and use that instead.
		machine, err := getMachineByNodeName(ctx, c, sshNode.NodeName)
		if err != nil {
			return sshNode, nil, fmt.Errorf("failed to find SSH key for node [%s]", nodeName)
		}

		link = machine.Links["sshkeys"]
	}

	key, sshUser, err := getSSHKey(c, link, getNodeName(sshNode))
	if err != nil {
		return sshNode, nil, err
	}
	if sshUser != "" {
		sshNode.SshUser = sshUser
	}

	return sshNode, key, nil
}

func callSSH(content []byte, ip string, user string, args []string) error {
	dest := fmt.Sprintf("%s@%s", user, ip)

	tmpfile, err := os.CreateTemp("", "ssh")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	if err := os.Chmod(tmpfile.Name(), 0600); err != nil {
		return err
	}

	_, err = tmpfile.Write(content)
	if err != nil {
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}

	cmd := exec.Command("ssh", append([]string{"-i", tmpfile.Name(), dest}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getSSHKey(c *cliclient.MasterClient, link, nodeName string) ([]byte, string, error) {
	if link == "" {
		return nil, "", fmt.Errorf("failed to find SSH key for %s", nodeName)
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, "", err
	}
	req.SetBasicAuth(c.UserConfig.AccessKey, c.UserConfig.SecretKey)
	req.Header.Add("Accept-Encoding", "zip")

	client := &http.Client{}

	if c.UserConfig.CACerts != "" {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(c.UserConfig.CACerts))
		if !ok {
			return []byte{}, "", err
		}
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: roots,
			},
		}
		client.Transport = tr
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	zipFiles, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("%s", zipFiles)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipFiles), resp.ContentLength)
	if err != nil {
		return nil, "", err
	}

	var sshKey []byte
	var sshUser string
	for _, file := range zipReader.File {
		if path.Base(file.Name) == "id_rsa" {
			sshKey, err = readFile(file)
			if err != nil {
				return nil, "", err
			}
		} else if path.Base(file.Name) == "config.json" {
			config, err := readFile(file)
			if err != nil {
				return nil, "", err
			}

			var data map[string]interface{}
			err = json.Unmarshal(config, &data)
			if err != nil {
				return nil, "", err
			}
			sshUser, _ = data["SSHUser"].(string)
		}
	}
	if len(sshKey) == 0 {
		return sshKey, "", errors.New("can't find private key file")
	}
	return sshKey, sshUser, nil
}

func readFile(file *zip.File) ([]byte, error) {
	r, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
