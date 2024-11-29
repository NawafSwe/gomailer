package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage_EncodeBase64(t *testing.T) {
	t.Parallel()
	t.Run("should encode message to base64", func(t *testing.T) {
		t.Parallel()
		expected := "aW5wdXQ"
		msg := encodeBase64("input")
		assert.Equal(t, expected, msg)
	})
}

func TestMessage_SplitLines(t *testing.T) {
	input := "input"
	t.Run("should put input into multiple lines when it is exceeding the max length", func(t *testing.T) {
		lines := splitLines(input, 1)
		assert.Equal(t, len(lines), len(input))
		// assert each line.
		for i, line := range lines {
			assert.Equal(t, input[i:i+1], line)
		}

	})

	t.Run("should put input in one line when it is not exceeding the max length", func(t *testing.T) {

		lines := splitLines(input, 20)
		assert.Equal(t, len(lines), 1)
		assert.Equal(t, lines[0], input)
	})
}

func TestMessage_Encode(t *testing.T) {
	// testing message with html body and normal text
	// testing message with body only
	// testing message with html only
	// testing message with html body attachments
	// testing with attachment only

}
