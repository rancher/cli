package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExternalIDInVersion(t *testing.T) {
	assert := assert.New(t)

	got, err := updateExternalIDVersion("catalog://?catalog=library&template=cert-manager&version=v0.5.2", "v1.2.3")
	assert.Nil(err)
	assert.Equal("catalog://?catalog=library&template=cert-manager&version=v1.2.3", got)

	got, err = updateExternalIDVersion("catalog://?catalog=c-29wkq/clusterscope&type=clusterCatalog&template=mysql&version=0.3.8", "0.3.9")
	assert.Nil(err)
	assert.Equal("catalog://?catalog=c-29wkq/clusterscope&type=clusterCatalog&template=mysql&version=0.3.9", got)

	got, err = updateExternalIDVersion("catalog://?catalog=p-j9gfw/projectscope&type=projectCatalog&template=grafana&version=0.0.31", "0.0.30")
	assert.Nil(err)
	assert.Equal("catalog://?catalog=p-j9gfw/projectscope&type=projectCatalog&template=grafana&version=0.0.30", got)
}
