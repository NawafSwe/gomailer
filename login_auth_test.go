package gomailer

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/smtp"
	"testing"
)

func TestLoginAuth_Start(t *testing.T) {
	t.Run("should starts login auth successfully", func(t *testing.T) {
		login := newSmtpLoginAuth(testUser, testPassword)
		p, srvInfo, err := login.Start(&smtp.ServerInfo{Name: testLocalName})
		assert.Nil(t, err)
		assert.NotNil(t, srvInfo)
		assert.NotEmpty(t, p)
	})

}

func TestLoginAuth_Next(t *testing.T) {
	t.Run("should successfully call next for accepting username and password", func(t *testing.T) {
		t.Parallel()
		login := newSmtpLoginAuth(testUser, testPassword)
		usernameInfo, err := login.Next([]byte("Username:"), true)
		assert.Nil(t, err)
		assert.Equal(t, usernameInfo, []byte(testUser))

		passwordInfo, err := login.Next([]byte("Password:"), true)
		assert.Nil(t, err)
		assert.Equal(t, passwordInfo, []byte(testPassword))

	})
	t.Run("should fail to call next for accepting username and password due to unknown challenge", func(t *testing.T) {
		t.Parallel()
		login := newSmtpLoginAuth(testUser, testPassword)
		challenge := "Unknown:"
		info, err := login.Next([]byte(challenge), true)
		assert.NotNil(t, err)
		assert.Nil(t, info)
		assert.Equal(t, fmt.Errorf("unexpected server challange: %s", challenge), err)

	})
	t.Run("should return nil info when no more data there", func(t *testing.T) {
		t.Parallel()
		login := newSmtpLoginAuth(testUser, testPassword)
		info, err := login.Next([]byte(""), false)
		assert.Nil(t, err)
		assert.Nil(t, info)

	})
}
