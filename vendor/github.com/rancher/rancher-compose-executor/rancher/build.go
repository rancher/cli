package rancher

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/builder"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/rancher/rancher-compose-executor/project"
)

const DefaultDockerfileName = "Dockerfile"

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

	tar, err := createTar(service.Build.Context, service.Build.Dockerfile)
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

func createTar(contextDirectory, dockerfile string) (io.ReadCloser, error) {
	// This code was ripped off from docker/api/client/build.go
	dockerfileName := filepath.Join(contextDirectory, dockerfile)

	absContextDirectory, err := filepath.Abs(contextDirectory)
	if err != nil {
		return nil, err
	}

	filename := dockerfileName

	if dockerfile == "" {
		// No -f/--file was specified so use the default
		dockerfileName = DefaultDockerfileName
		filename = filepath.Join(absContextDirectory, dockerfileName)

		// Just to be nice ;-) look for 'dockerfile' too but only
		// use it if we found it, otherwise ignore this check
		if _, err = os.Lstat(filename); os.IsNotExist(err) {
			tmpFN := path.Join(absContextDirectory, strings.ToLower(dockerfileName))
			if _, err = os.Lstat(tmpFN); err == nil {
				dockerfileName = strings.ToLower(dockerfileName)
				filename = tmpFN
			}
		}
	}

	origDockerfile := dockerfileName // used for error msg
	if filename, err = filepath.Abs(filename); err != nil {
		return nil, err
	}

	// Now reset the dockerfileName to be relative to the build context
	dockerfileName, err = filepath.Rel(absContextDirectory, filename)
	if err != nil {
		return nil, err
	}

	// And canonicalize dockerfile name to a platform-independent one
	dockerfileName, err = archive.CanonicalTarNameForPath(dockerfileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot canonicalize dockerfile path %s: %v", dockerfileName, err)
	}

	if _, err = os.Lstat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cannot locate Dockerfile: %s", origDockerfile)
	}
	var includes = []string{"."}
	var excludes []string

	dockerIgnorePath := path.Join(contextDirectory, ".dockerignore")
	dockerIgnore, err := os.Open(dockerIgnorePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		logrus.Warnf("Error while reading .dockerignore (%s) : %s", dockerIgnorePath, err.Error())
		excludes = make([]string, 0)
	} else {
		excludes, err = dockerignore.ReadAll(dockerIgnore)
		if err != nil {
			return nil, err
		}
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed.  The deamon will remove them for us, if needed, after it
	// parses the Dockerfile.
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(dockerfileName, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", dockerfileName)
	}

	if err := builder.ValidateContextDirectory(contextDirectory, excludes); err != nil {
		return nil, fmt.Errorf("Error checking context is accessible: '%s'. Please check permissions and try again.", err)
	}

	options := &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	}

	return archive.TarWithOptions(contextDirectory, options)
}
