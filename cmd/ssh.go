package cmd

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/types/client/management/v3"
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
	$ rancher ssh login1@nodeFoo env
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
				Usage: "Use the external ip address",
			},
			cli.StringFlag{
				Name:  "login,l",
				Usage: "The login name",
			},
		},
		SkipFlagParsing: false,
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

	resource, err := Lookup(c, nodeName, "node")
	if nil != err {
		return err
	}

	sshNode, err := getNodeByID(ctx, c, resource.ID)
	if nil != err {
		return err
	}

	if user == "" {
		user = sshNode.SshUser
	}

	key, err := getSSHKey(c, sshNode)
	if err != nil {
		return err
	}

	ipAddress := sshNode.IPAddress
	if ctx.Bool("external") {
		ipAddress = sshNode.ExternalIPAddress
	}

	return processExitCode(callSSH(key, ipAddress, user, args))
}

func callSSH(content []byte, ip string, user string, args []string) error {
	dest := fmt.Sprintf("%s@%s", user, ip)

	tmpfile, err := ioutil.TempFile("", "ssh")
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

func getSSHKey(c *cliclient.MasterClient, node managementClient.Node) ([]byte, error) {
	link, ok := node.Links["nodeConfig"]
	if !ok {
		return nil, fmt.Errorf("failed to find SSH key for %s", getNodeName(node))
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.UserConfig.AccessKey, c.UserConfig.SecretKey)
	req.Header.Add("Accept-Encoding", "zip")

	client := &http.Client{}

	if c.UserConfig.CACerts != "" {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(c.UserConfig.CACerts))
		if !ok {
			return []byte{}, err
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
		return nil, err
	}
	defer resp.Body.Close()

	zipFiles, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", zipFiles)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipFiles), resp.ContentLength)
	if err != nil {
		return nil, err
	}

	for _, file := range zipReader.File {
		if path.Base(file.Name) == "id_rsa" {
			r, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer r.Close()
			return ioutil.ReadAll(r)
		}
	}
	return nil, errors.New("can't find private key file")
}
