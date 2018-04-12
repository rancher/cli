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

	"github.com/pkg/errors"
	"github.com/rancher/cli/cliclient"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/urfave/cli"
)

const sshDescription = `For any nodes created through Rancher using docker-machine, 
you can SSH into the node. This is not supported for any custom nodes.`

func SSHCommand() cli.Command {
	return cli.Command{
		Name:            "ssh",
		Usage:           "SSH into a node",
		Description:     sshDescription,
		ArgsUsage:       "[NODEID NODENAME]",
		Action:          nodeSSH,
		Flags:           []cli.Flag{},
		SkipFlagParsing: true,
	}
}

func nodeSSH(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return errors.New("node ID is required")
	}
	args := ctx.Args()
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return cli.ShowCommandHelp(ctx, "ssh")
	}

	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	resource, err := Lookup(c, args[0], "node")
	if nil != err {
		return err
	}

	sshNode, err := getNodeByID(ctx, c, resource.ID)
	if nil != err {
		return err
	}

	key, err := getSSHKey(c, sshNode)
	if err != nil {
		return err
	}

	return processExitCode(callSSH(key, sshNode.IPAddress, sshNode.SshUser))
}

func callSSH(content []byte, ip string, user string) error {
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

	cmd := exec.Command("ssh", append([]string{"-i", tmpfile.Name()}, dest)...)
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
