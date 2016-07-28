package rancher

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/docker/builder"
	"github.com/docker/libcompose/project"
)

type Uploader interface {
	Upload(p *project.Project, name string, reader io.ReadSeeker, hash string) (string, string, error)
	Name() string
}

func Upload(c *Context, name string) (string, string, error) {
	uploader := c.Uploader
	if uploader == nil {
		return "", "", errors.New("Build not supported")
	}
	p := c.Project
	logrus.Infof("Uploading build for %s using provider %s", name, uploader.Name())

	content, hash, err := createBuildArchive(p, name)
	if err != nil {
		return "", "", err
	}

	return uploader.Upload(p, name, content, hash)
}

func createBuildArchive(p *project.Project, name string) (io.ReadSeeker, string, error) {
	service, ok := p.ServiceConfigs.Get(name)
	if !ok {
		return nil, "", fmt.Errorf("No such service: %s", name)
	}

	tar, err := builder.CreateTar(service.Build.Context, service.Build.Dockerfile)
	if err != nil {
		return nil, "", err
	}
	defer tar.Close()

	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, "", err
	}

	if err := os.Remove(tempFile.Name()); err != nil {
		tempFile.Close()
		return nil, "", err
	}

	digest := sha256.New()
	output := io.MultiWriter(tempFile, digest)

	_, err = io.Copy(output, tar)
	if err != nil {
		tempFile.Close()
		return nil, "", err
	}

	hexString := hex.EncodeToString(digest.Sum([]byte{}))
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		tempFile.Close()
		return nil, "", err
	}

	return tempFile, hexString, nil
}
