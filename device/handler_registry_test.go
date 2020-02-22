package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlers(t *testing.T) {
	// Add one
	handler := &mockHandler{name: "mockHandler"}
	MustAddHandler(handler)
	// Add it one more time
	assert.Panics(t, func() {
		MustAddHandler(handler)
	})

	// Lookup it
	assert.NotNil(t, FindHandlerByName(handler.GetName()))
	// Lookup non existing device handler
	assert.Nil(t, FindHandlerByName("test111fsdfsd"))
}
