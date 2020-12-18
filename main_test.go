package main

import (
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMain(t *testing.T) {
}

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	plugin.Tags = []string{"tag1", "tag2"}
	plugin.All = true
	i, e := checkArgs(event)
	assert.Error(e)
	assert.Equal(sensu.CheckStateCritical, i)
	plugin.All = false
	i, e = checkArgs(event)
	assert.NoError(e)
	assert.Equal(sensu.CheckStateOK, i)
}
