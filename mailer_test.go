package gomailer

import (
	"crypto/tls"
	"fmt"
	mailerMock "github.com/NawafSwe/gomailer/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"net/smtp"
	"testing"
	"time"
)

// test variables.
const (
	testPort      = 587
	testSSLPort   = 465
	testUser      = "test_user"
	testPassword  = "test@123"
	testHost      = "localhost.smtp.com"
	testLocalName = "local.goMailer"
)

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
				WithTLSConfig(&tls.Config{ServerName: testHost}),
				WithAuth(smtp.PlainAuth("", testUser, testPassword, testHost)),
			},
		}
		mailer := NewMailer(arg.host, arg.port, arg.username, arg.password, arg.options...)
		assert.NotNil(t, mailer)
	})

}

func TestMailer_Dial(t *testing.T) {
	dummyErr := fmt.Errorf("dummy error")
	t.Run("should dial smtp server via mailer without tls config using plain auth", func(t *testing.T) {
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
		smtpSender, err := mailer.Dial()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)
	})
	t.Run("should dial smtp server via mailer without tls config using login auth", func(t *testing.T) {
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
		smtpSender, err := mailer.Dial()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)
	})
	t.Run("should dial smtp server using ssl connection with CRAM-MD5 auth mechanism", func(t *testing.T) {
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
		mailer := NewMailer(testHost, testSSLPort, testUser, testPassword)
		assert.NotNil(t, mailer)

		// expect on mocks
		smtpMock.EXPECT().Extension("AUTH").Return(true, crmAuthMechanism)
		smtpMock.EXPECT().Auth(authMock).Return(nil)

		// dial smtp server and obtain sender.
		smtpSender, err := mailer.Dial()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)
	})
	t.Run("should dial smtp server with STARTTLS and plain auth when localName is specified", func(t *testing.T) {
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
		smtpSender, err := mailer.Dial()

		assert.Nil(t, err)
		assert.NotNil(t, smtpSender)
	})
	t.Run("should fail to dial smtp server when failed to establish a tcp connection", func(t *testing.T) {

		netDialTimeout = func(network string, host string, t time.Duration) (net.Conn, error) {
			return nil, dummyErr
		}

		mailer := NewMailer(testHost, testPort, testUser, testPassword)
		smtpSender, err := mailer.Dial()
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to dial to smtp server: %w", dummyErr), err)
		assert.Nil(t, smtpSender)

	})
	t.Run("should fail to dial smtp server when failed to create a smtp client", func(t *testing.T) {

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
		smtpSender, err := mailer.Dial()
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to dial smtp server: %w", dummyErr), err)
		assert.Nil(t, smtpSender)

	})
	t.Run("should fail to dial smtp server when failed to issue HELLO cmd to smtp server", func(t *testing.T) {
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
		smtpSender, err := mailer.Dial()
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to dial smtp server: %w", dummyErr), err)
		assert.Nil(t, smtpSender)

	})
	t.Run("should fail dial smtp server when server received a failure in issuing STARTTLS cmd", func(t *testing.T) {
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
		smtpSender, err := mailer.Dial()

		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to StartTLS: %w", dummyErr), err)
		assert.Nil(t, smtpSender)
	})
	t.Run("should fail dial smtp server via mailer using tls config when smtp failed to authenticate with smtp server", func(t *testing.T) {
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
		smtpSender, err := mailer.Dial()

		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to authenticate with smtp server: %w", dummyErr), err)
		assert.Nil(t, smtpSender)
	})
	// fail hello
	// fail net dail timeout
	// fail create smtp client
	// fail StartTLS (asser conn close called and return no error)
	// enable logging option
	// support plain login auth/ test unsupported method.
	// fail Auth

}
