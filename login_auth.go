package gomailer

import (
	"fmt"
	"net/smtp"
)

const (
	loginAuthUsername = "Username:"
	loginAuthPassword = "Password:"
)

// loginAuth implements the smtp.Auth interface for LOGIN authentication mechanism.
type loginAuth struct {
	username, password string
}

// Start begins the LOGIN authentication with the server.
func (a *loginAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	return loginAuthMechanism, []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case loginAuthUsername:
			return []byte(a.username), nil
		case loginAuthPassword:
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unexpected server challange: %s", fromServer)
		}
	}
	return nil, nil

}

// newSmtpLoginAuth returns a new loginAuth.
func newSmtpLoginAuth(username, password string) auth {
	return &loginAuth{username: username, password: password}
}
