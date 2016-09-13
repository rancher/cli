package lookup

import (
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/lookup"
)

type FileResourceLookup struct {
	lookup.FileResourceLookup
}

// Give a warning rather than resolve relative paths
func (f *FileResourceLookup) ResolvePath(path, inFile string) string {
	vs := strings.SplitN(path, ":", 2)
	if len(vs) == 2 && !filepath.IsAbs(vs[0]) {
		log.Warnf("Rancher Compose will not resolve relative path %s", vs[0])
	}
	return path
}
