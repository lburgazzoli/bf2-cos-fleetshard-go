package v2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRawMessage(t *testing.T) {

	t.Run("by value", func(t *testing.T) {
		var msg RawMessage

		err := msg.Set(KafkaSpec{
			ID:  "foo",
			URL: "kafka.acme.com:443",
		})

		assert.Nil(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, RawMessage(`{"id":"foo","url":"kafka.acme.com:443"}`), msg)
	})

	t.Run("by address", func(t *testing.T) {
		var msg RawMessage

		err := msg.Set(&KafkaSpec{
			ID:  "foo",
			URL: "kafka.acme.com:443",
		})

		assert.Nil(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, RawMessage(`{"id":"foo","url":"kafka.acme.com:443"}`), msg)
	})
}
