package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher/client"
)

type Config struct {
	AccessKey   string `json:"accessKey"`
	SecretKey   string `json:"secretKey"`
	URL         string `json:"url"`
	Environment string `json:"environment"`
	Path        string `json:"path,omitempty"`
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
		Name:   "config",
		Usage:  "Setup client configuration",
		Action: errorWrapper(configSetup),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "dump",
				Usage: "Dump the effective configuration",
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
	if err != nil {
		return err
	}

	if ctx.Bool("dump") {
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

	project, err := GetEnvironment(config.Environment, c)
	if err != errNoEnv {
		if err != nil {
			return err
		}
		config.Environment = project.Id
	}

	return config.Write()
}
