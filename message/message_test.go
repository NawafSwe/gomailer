package message

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage(t *testing.T) {

	tests := map[string]struct {
		getMessage  func() Message
		expectedErr error
	}{
		"should successfully encode message when message constructed with valid from and recipients": {
			getMessage: func() Message {
				msg := NewMessage()
				msg.From = testEmail
				msg.Recipients = []string{testEmail}
				return msg
			},
		},
		"should fail encoding message when no from address provided": {
			getMessage: func() Message {
				return NewMessage()
			},
			expectedErr: fmt.Errorf("failed to encode message: %w", fmt.Errorf("from address cannot be empty")),
		},
		"should fail encoding message when invalid from address provided": {
			getMessage: func() Message {
				msg := NewMessage()
				msg.From = "invalid"
				return msg
			},
			expectedErr: fmt.Errorf("failed to encode message: %w", fmt.Errorf("invalid from address: %w", fmt.Errorf("mail: missing '@' or angle-addr"))),
		},
		"should fail encoding message when no recipients address provided": {
			getMessage: func() Message {
				msg := NewMessage()
				msg.From = testEmail
				return msg
			},
			expectedErr: fmt.Errorf("failed to encode message: %w", fmt.Errorf("recipients cannot be empty slice")),
		},
		"should fail encoding message when invalid recipients address provided": {
			getMessage: func() Message {
				msg := NewMessage()
				msg.From = testEmail
				msg.Recipients = []string{"gomailerAddr"}
				return msg
			},
			expectedErr: fmt.Errorf("failed to encode message: %w", fmt.Errorf("given gomailerAddr is invalid recipient email: %w", fmt.Errorf("mail: missing '@' or angle-addr"))),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.getMessage != nil {
				_, err := tc.getMessage().Encode()
				assert.Equal(t, tc.expectedErr, err)
			}

		})
	}
}
