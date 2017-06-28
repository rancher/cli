package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/rancher/go-rancher/v2"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

type Config struct {
	AccessKey   string `json:"accessKey"`
	SecretKey   string `json:"secretKey"`
	URL         string `json:"url"`
	Environment string `json:"environment"`
	Path        string `json:"path,omitempty"`
}

func baseURL(fullURL string) (string, error) {
	idx := strings.LastIndex(fullURL, "/v2-beta")
	if idx == -1 {
		u, err := url.Parse(fullURL)
		if err != nil {
			return "", err
		}
		newURL := url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
		}
		return newURL.String(), nil
	}
	return fullURL[:idx], nil
}

func (c Config) EnvironmentURL() (string, error) {
	projectID := c.Environment
	if projectID == "" || !strings.HasPrefix(projectID, "1a") {
		rancherClient, err := client.NewRancherClient(&client.ClientOpts{
			Url:       c.URL,
			AccessKey: c.AccessKey,
			SecretKey: c.SecretKey,
		})
		if err != nil {
			return "", err
		}
		project, err := GetEnvironment(c.Environment, rancherClient)
		if err != nil {
			return "", err
		}
		if project != nil {
			projectID = project.Id
		}
	}

	url, err := baseURL(c.URL)
	if err != nil {
		return "", err
	}
	if projectID == "" || projectID == "ALL" {
		url = url + "/v2-beta/schemas"
	} else {
		url = url + "/v2-beta/projects/" + projectID + "/schemas"
	}
	return url, nil
}

func (c Config) Write() error {
	err := os.MkdirAll(path.Dir(c.Path), 0700)
	if err != nil {
		return err
	}

	logrus.Infof("Saving config to %s", c.Path)
	p := c.Path
	c.Path = ""
	output, err := os.Create(p)
	if err != nil {
		return err
	}
	defer output.Close()

	return json.NewEncoder(output).Encode(c)
}

func LoadConfig(path string) (Config, error) {
	config := Config{
		Path: path,
	}

	content, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return config, nil
	} else if err != nil {
		return config, err
	}

	err = json.Unmarshal(content, &config)
	config.Path = path

	return config, err
}

func ConfigCommand() cli.Command {
	return cli.Command{
		Name:      "config",
		Usage:     "Setup client configuration",
		Action:    configSetup,
		ArgsUsage: "None",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "print,p",
				Usage: "Print the current configuration",
			},
		},
	}
}

func getConfig(reader *bufio.Reader, text, def string) (string, error) {
	for {
		fmt.Printf("%s [%s]: ", text, def)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)

		if input != "" {
			return input, nil
		}

		if input == "" && def != "" {
			return def, nil
		}
	}
}

func configSetup(ctx *cli.Context) error {
	config, err := lookupConfig(ctx)
	if err != nil && err != errNoURL {
		return err
	}

	if ctx.Bool("print") {
		return json.NewEncoder(os.Stdout).Encode(config)
	}

	reader := bufio.NewReader(os.Stdin)

	config.URL, err = getConfig(reader, "URL", config.URL)
	if err != nil {
		return err
	}

	config.AccessKey, err = getConfig(reader, "Access Key", config.AccessKey)
	if err != nil {
		return err
	}

	config.SecretKey, err = getConfig(reader, "Secret Key", config.SecretKey)
	if err != nil {
		return err
	}

	c, err := client.NewRancherClient(&client.ClientOpts{
		Url:       config.URL,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	})
	if err != nil {
		return err
	}

	if schema, ok := c.GetSchemas().CheckSchema("schema"); ok {
		// Normalize URL
		config.URL = schema.Links["collection"]
	} else {
		return fmt.Errorf("Failed to find schema URL")
	}

	c, err = client.NewRancherClient(&client.ClientOpts{
		Url:       config.URL,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	})
	if err != nil {
		return err
	}

	project, err := GetEnvironment("", c)
	if err != errNoEnv {
		if err != nil {
			return err
		}
		config.Environment = project.Id
	}

	return config.Write()
}
