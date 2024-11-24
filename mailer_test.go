package gomailer

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"testing"
	"time"

	mailerMock "github.com/NawafSwe/gomailer/internal/mock"
	"github.com/NawafSwe/gomailer/message"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// test variables.

const (
	testPort      = 587
	testSSLPort   = 465
	testUser      = "test_user"
	testPassword  = "test@123"
	testHost      = "localhost.smtp.com"
	testLocalName = "localName.goMailer"
	testFromEmail = "test@gomailer.com"
)

var testRecipient = []string{"test@gomailer.com"}

func TestMailer_NewMailer(t *testing.T) {
	// Possible SMTP-client stuff for iteration with mock server
	type args struct {
		port      int
		host      string
		username  string
		password  string
		localName string
		options   []Options
		auth      smtp.Auth
		tlsConfig *tls.Config
	}

	t.Run("should successfully creates mailer without config", func(t *testing.T) {
		t.Parallel()
		arg := args{
			port:      testPort,
			host:      testHost,
			username:  testUser,
			password:  testPassword,
			localName: testLocalName,
		}
		mailer := NewMailer(arg.host, arg.port, arg.username, arg.password, arg.options...)
		assert.NotNil(t, mailer)
	})

	t.Run("should successfully creates mailer with all possible configs", func(t *testing.T) {
		t.Parallel()
		arg := args{
			port:     testPort,
			host:     testHost,
			username: testUser,
			password: testPassword,
			options: []Options{
				WithLocalName(testLocalName),
				WithDialTimeout(time.Second),
				WithSecrets(""),
				WithSSLEnabled(true),
				WithTLSConfig(&tls.Config{ServerName: testHost}),
				WithAuth(smtp.PlainAuth("", testUser, testPassword, testHost)),
			},
		}
		mailer := NewMailer(arg.host, arg.port, arg.username, arg.password, arg.options...)
		assert.NotNil(t, mailer)
	})
}

func TestMailer_ConnectAndAuthenticate(t *testing.T) {
	dummyErr := fmt.Errorf("dummy error")
	t.Run("should connect and authenticate to smtp server via mailer without tls config using plain auth", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpPlainAuth = func(identity, username, password, host string) auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testPort, testUser, testPassword)
		assert.NotNil(t, mailer)

		// expect on mocks
		smtpMock.EXPECT().Extension("STARTTLS").Return(false, "STARTTLS")
		smtpMock.EXPECT().Extension("AUTH").Return(true, plainAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)
	})
	t.Run("should connect and authenticate to smtp server via mailer without tls config using login auth", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		// init mailer
		mailer := NewMailer(testHost, testPort, testUser, testPassword)
		assert.NotNil(t, mailer)

		// expect on mocks
		smtpMock.EXPECT().Extension("STARTTLS").Return(false, "STARTTLS")
		smtpMock.EXPECT().Extension("AUTH").Return(true, loginAuthMechanism)
		smtpMock.EXPECT().Auth(newSmtpLoginAuth(testUser, testPassword)).Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)
	})
	t.Run("should connect and authenticate to smtp server using ssl connection with CRAM-MD5 auth mechanism", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)

		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)
	})
	t.Run("should connect and authenticate to smtp server with STARTTLS and plain auth when localName is specified", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}

		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpPlainAuth = func(identity, username, password, host string) auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testPort, testUser, testPassword, WithLocalName(testLocalName))
		assert.NotNil(t, mailer)

		// expect on mocks
		smtpMock.EXPECT().Hello(testLocalName).Return(nil)
		smtpMock.EXPECT().Extension("STARTTLS").Return(true, "STARTTLS")
		smtpMock.EXPECT().StartTLS(mailer.tlsConfig).Return(nil)
		smtpMock.EXPECT().Extension("AUTH").Return(true, plainAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)
	})
	t.Run("should fail to connect and authenticate to smtp server when failed to establish a tcp connection", func(t *testing.T) {
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return nil, dummyErr
		}

		mailer := NewMailer(testHost, testPort, testUser, testPassword)
		smtpSender, err := mailer.ConnectAndAuthenticate()
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to dial to smtp server: %w", dummyErr), err)
		assert.Nil(t, smtpSender)
	})
	t.Run("should fail to connect and authenticate to smtp server when failed to create a smtp client", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		netConnMock := mailerMock.NewMockconn(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return nil, dummyErr
		}

		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		mailer := NewMailer(testHost, testPort, testUser, testPassword)
		smtpSender, err := mailer.ConnectAndAuthenticate()
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to dial smtp server: %w", dummyErr), err)
		assert.Nil(t, smtpSender)
	})
	t.Run("should fail to connect and authenticate to SMTP server when issuing HELLO command fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}

		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		// init mailer
		mailer := NewMailer(testHost, testPort, testUser, testPassword, WithLocalName(testLocalName))
		assert.NotNil(t, mailer)

		// expect on mocks
		smtpMock.EXPECT().Hello(testLocalName).Return(dummyErr)
		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to dial smtp server: %w", dummyErr), err)
		assert.Nil(t, smtpSender)
	})
	t.Run("should fail to connect and authenticate to SMTP server when issuing STARTTLS command fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}

		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		// init mailer
		mailer := NewMailer(testHost, testPort, testUser, testPassword, WithLocalName(testLocalName))
		assert.NotNil(t, mailer)

		// expect on mocks
		smtpMock.EXPECT().Hello(testLocalName).Return(nil)
		smtpMock.EXPECT().Extension("STARTTLS").Return(true, "STARTTLS")
		smtpMock.EXPECT().StartTLS(mailer.tlsConfig).Return(dummyErr)
		smtpMock.EXPECT().Close().Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to StartTLS: %w", dummyErr), err)
		assert.Nil(t, smtpSender)
	})
	t.Run("should fail connect and authenticate to smtp server via mailer using tls config when smtp failed to authenticate with smtp server", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpPlainAuth = func(identity, username, password, host string) auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testPort, testUser, testPassword)
		assert.NotNil(t, mailer)

		// expect on mocks
		smtpMock.EXPECT().Extension("STARTTLS").Return(false, "STARTTLS")
		smtpMock.EXPECT().Extension("AUTH").Return(true, plainAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(dummyErr)
		smtpMock.EXPECT().Close().Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to authenticate with smtp server: %w", dummyErr), err)
		assert.Nil(t, smtpSender)
	})
}

func TestMailer_Send(t *testing.T) {
	dummyErr := fmt.Errorf("dummy error")
	t.Run("should send message successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)
		writeCloserMock := mailerMock.NewMockwriteCloser(ctrl)
		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)

		msg := message.Message{
			From:       testFromEmail,
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)
		smtpMock.EXPECT().Mail(msg.From).Return(nil)
		smtpMock.EXPECT().Rcpt(msg.Recipients[0]).Return(nil)
		smtpMock.EXPECT().Data().Return(writeCloserMock, nil)
		smtpMock.EXPECT().Quit().Return(nil)
		writeCloserMock.EXPECT().Write(gomock.Any()).Return(0, nil)
		writeCloserMock.EXPECT().Close().Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)

		err = smtpSender.Send(msg)
		assert.Nil(t, err)
		assert.Nil(t, smtpSender.Close())
	})
	t.Run("should success send message without using mailSender implementation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)
		writeCloserMock := mailerMock.NewMockwriteCloser(ctrl)
		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)

		msg := message.Message{
			From:       testFromEmail,
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)
		smtpMock.EXPECT().Mail(msg.From).Return(nil)
		smtpMock.EXPECT().Rcpt(msg.Recipients[0]).Return(nil)
		smtpMock.EXPECT().Data().Return(writeCloserMock, nil)
		smtpMock.EXPECT().Quit().Return(nil)
		writeCloserMock.EXPECT().Write(gomock.Any()).Return(0, nil)
		writeCloserMock.EXPECT().Close().Return(nil)

		// dial smtp server and obtain sender.
		err := mailer.Send(msg)
		assert.Nil(t, err)
	})
	t.Run("should send message successfully and failed in terminating the session", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)
		writeCloserMock := mailerMock.NewMockwriteCloser(ctrl)
		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)

		msg := message.Message{
			From:       testFromEmail,
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)
		smtpMock.EXPECT().Mail(msg.From).Return(nil)
		smtpMock.EXPECT().Rcpt(msg.Recipients[0]).Return(nil)
		smtpMock.EXPECT().Data().Return(writeCloserMock, nil)
		smtpMock.EXPECT().Quit().Return(dummyErr)
		writeCloserMock.EXPECT().Write(gomock.Any()).Return(0, nil)
		writeCloserMock.EXPECT().Close().Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)

		err = smtpSender.Send(msg)
		assert.Nil(t, err)
		err = smtpSender.Close()
		assert.Equal(t, fmt.Errorf("failed to close connection to smtp server: %w", dummyErr), err)
	})
	t.Run("should fail to send message when issuing MAIL command fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)
		msg := message.Message{
			From:       testFromEmail,
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)
		smtpMock.EXPECT().Mail(msg.From).Return(dummyErr)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)

		err = smtpSender.Send(msg)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("mailer failed to send MAIL command for address %s: %w", msg.From, dummyErr), err)
	})
	t.Run("should fail to send message when issuing RCPT command fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)
		msg := message.Message{
			From:       testFromEmail,
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)
		smtpMock.EXPECT().Mail(msg.From).Return(nil)
		smtpMock.EXPECT().Rcpt(msg.Recipients[0]).Return(dummyErr)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)

		err = smtpSender.Send(msg)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("mailer failed to send rcpt command for address %s: %w", msg.Recipients[0], dummyErr), err)
	})
	t.Run("should fail to send message when getting writer closer from SMTP client fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)
		writeCloserMock := mailerMock.NewMockwriteCloser(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)
		msg := message.Message{
			From:       testFromEmail,
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)
		smtpMock.EXPECT().Mail(msg.From).Return(nil)
		smtpMock.EXPECT().Rcpt(msg.Recipients[0]).Return(nil)
		smtpMock.EXPECT().Data().Return(writeCloserMock, dummyErr)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)

		err = smtpSender.Send(msg)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("mailer failed to get data writer: %w", dummyErr), err)
	})
	t.Run("should fail to send message when encoding message fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)
		writeCloserMock := mailerMock.NewMockwriteCloser(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)
		msg := message.Message{
			From:       "",
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)
		smtpMock.EXPECT().Mail(msg.From).Return(nil)
		smtpMock.EXPECT().Rcpt(msg.Recipients[0]).Return(nil)
		smtpMock.EXPECT().Data().Return(writeCloserMock, nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)

		err = smtpSender.Send(msg)
		assert.NotNil(t, err)
		assert.Equal(t, "failed to send message: failed to encode message: from address cannot be empty", err.Error())
	})
	t.Run("should fail to send message when writing encoded message fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)
		writeCloserMock := mailerMock.NewMockwriteCloser(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)
		msg := message.Message{
			From:       testFromEmail,
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)
		smtpMock.EXPECT().Mail(msg.From).Return(nil)
		smtpMock.EXPECT().Rcpt(msg.Recipients[0]).Return(nil)
		smtpMock.EXPECT().Data().Return(writeCloserMock, nil)
		writeCloserMock.EXPECT().Write(gomock.Any()).Return(0, dummyErr)
		writeCloserMock.EXPECT().Close().Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.ConnectAndAuthenticate()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)

		err = smtpSender.Send(msg)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed writing data: %w", dummyErr), err)
	})
	t.Run("should fail to send message due to authentication failure without using mailSender implementation", func(t *testing.T) {
		// stub functions
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return nil, dummyErr
		}
		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, testPassword)
		assert.NotNil(t, mailer)

		msg := message.Message{
			From:       testFromEmail,
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks

		// dial smtp server and obtain sender.
		err := mailer.Send(msg)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to connect and authenticate: %w", fmt.Errorf("failed to dial to smtp server: %w", dummyErr)), err)
	})
	t.Run("should fail to send message due to message sending failure without using mailSender implementation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		// prepare mocks
		smtpMock := mailerMock.NewMocksmtpClient(ctrl)
		netConnMock := mailerMock.NewMockconn(ctrl)
		authMock := mailerMock.NewMockauth(ctrl)
		writeCloserMock := mailerMock.NewMockwriteCloser(ctrl)

		// stub functions
		newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
			return smtpMock, nil
		}
		tlsClient = func(conn net.Conn, config *tls.Config) *tls.Conn {
			return &tls.Conn{}
		}
		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return netConnMock, nil
		}

		smtpCRAMMD5Auth = func(username, secret string) smtp.Auth {
			return authMock
		}

		// init mailer
		mailer := NewMailer(testHost, testSSLPort, testUser, "", WithSSLEnabled(true), WithSecrets(testPassword))
		assert.NotNil(t, mailer)
		msg := message.Message{
			From:       "",
			Recipients: testRecipient,
			Body:       "dummy body",
		}
		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)
		smtpMock.EXPECT().Mail(msg.From).Return(nil)
		smtpMock.EXPECT().Rcpt(msg.Recipients[0]).Return(nil)
		smtpMock.EXPECT().Data().Return(writeCloserMock, nil)
		smtpMock.EXPECT().Quit().Return(nil)

		err := mailer.Send(msg)
		assert.NotNil(t, err)
		assert.Equal(t, "failed to send message: failed to send message: failed to encode message: from address cannot be empty", err.Error())
	})
}
