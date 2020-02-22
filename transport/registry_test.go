package transport

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	transportByName = map[string]Transport{}
	t1 := newMockTransport("t1")

	// Add transport twice (+dup)
	err := AddTransport(t1)
	assert.NoError(t, err)
	err = AddTransport(t1)
	assert.Error(t, err)

	// Lookups
	assert.NotNil(t, FindTransportByName("t1"))

	// Delete it twice (again, dup)
	err = DeleteTransport("t1")
	assert.NoError(t, err)
	err = DeleteTransport("t1")
	assert.Error(t, err)

	// Ensure that it actually gets deleted
	assert.Nil(t, FindTransportByName("t1"))
}

var testConfig = `mock:
  transport11:
    str: transport string 11
    int: 11
  transport12:
    str: transport string 12
    int: 12
`

func TestRegistryLoadSave(t *testing.T) {
	transportByName = map[string]Transport{}
	transportTypes = map[string]transportCreateFunc{}

	// Register 2 transport types (used in config above)
	MustAddTransportType("mock", newMockTransport)

	// "Load" transports from YAML
	reader := bytes.NewReader([]byte(testConfig))
	require.NoError(t, LoadTransports(reader))

	// Save it back to YAML
	var writer bytes.Buffer
	assert.NoError(t, SaveTransports(&writer))

	assert.Equal(t, testConfig, writer.String())
}
