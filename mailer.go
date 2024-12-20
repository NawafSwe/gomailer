package gomailer

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/NawafSwe/gomailer/message"
)

const (
	sslPort            = 465
	crmAuthMechanism   = "CRAM-MD5"
	plainAuthMechanism = "PLAIN"
	loginAuthMechanism = "LOGIN"
)

//go:generate mockgen -source=mailer.go -destination=internal/mock/mailer.go -package=mock
type (
	// Options to configure Mailer.
	Options func(*Mailer)

	// auth is implemented by an SMTP authentication mechanism.
	auth interface {
		// Start begins an authentication with a server.
		// It returns the name of the authentication protocol
		// and optionally data to include in the initial AUTH message
		// sent to the server.
		// If it returns a non-nil error, the SMTP client aborts
		// the authentication attempt and closes the connection.
		Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error)

		// Next continues the authentication. The server has just sent
		// the fromServer data. If more is true, the server expects a
		// response, which Next should return as toServer; otherwise
		// Next should return toServer == nil.
		// If Next returns a non-nil error, the SMTP client aborts
		// the authentication attempt and closes the connection.
		Next(fromServer []byte, more bool) (toServer []byte, err error)
	}

	// smtpClient for implementing smtpClient
	smtpClient interface {
		Hello(string) error
		Extension(string) (bool, string)
		StartTLS(*tls.Config) error
		Auth(smtp.Auth) error
		Mail(string) error
		Rcpt(string) error
		Data() (io.WriteCloser, error)
		Quit() error
		Close() error
	}

	// SendCloser is an interface that encapsulates the functionality of sending a message and closing the connection to the SMTP server.
	// It provides methods to send an email message and to terminate the SMTP server session.
	SendCloser interface {
		// Close terminate the smtp server session.
		Close() error
		// Send sends message.Message.
		Send(message message.Message) error
	}

	// conn is a generic stream-oriented network connection.
	//
	// Multiple goroutines may invoke methods on a Conn simultaneously.
	conn interface {
		// Read reads data from the connection.
		// Read can be made to time out and return an error after a fixed
		// time limit; see SetDeadline and SetReadDeadline.
		Read(b []byte) (n int, err error)

		// Write writes data to the connection.
		// Write can be made to time out and return an error after a fixed
		// time limit; see SetDeadline and SetWriteDeadline.
		Write(b []byte) (n int, err error)

		// Close closes the connection.
		// Any blocked Read or Write operations will be unblocked and return errors.
		Close() error

		// LocalAddr returns the local network address, if known.
		LocalAddr() net.Addr

		// RemoteAddr returns the remote network address, if known.
		RemoteAddr() net.Addr

		// SetDeadline sets the read and write deadlines associated
		// with the connection. It is equivalent to calling both
		// SetReadDeadline and SetWriteDeadline.
		//
		// A deadline is an absolute time after which I/O operations
		// fail instead of blocking. The deadline applies to all future
		// and pending I/O, not just the immediately following call to
		// Read or Write. After a deadline has been exceeded, the
		// connection can be refreshed by setting a deadline in the future.
		//
		// If the deadline is exceeded a call to Read or Write or to other
		// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
		// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
		// The error's Timeout method will return true, but note that there
		// are other possible errors for which the Timeout method will
		// return true even if the deadline has not been exceeded.
		//
		// An idle timeout can be implemented by repeatedly extending
		// the deadline after successful Read or Write calls.
		//
		// A zero value for t means I/O operations will not time out.
		SetDeadline(t time.Time) error

		// SetReadDeadline sets the deadline for future Read calls
		// and any currently-blocked Read call.
		// A zero value for t means Read will not time out.
		SetReadDeadline(t time.Time) error

		// SetWriteDeadline sets the deadline for future Write calls
		// and any currently-blocked Write call.
		// Even if write times out, it may return n > 0, indicating that
		// some of the data was successfully written.
		// A zero value for t means Write will not time out.
		SetWriteDeadline(t time.Time) error
	}

	// writeCloser is an interface used to mock the smtp.Data writeCloser.
	// It encapsulates the methods required to write data to an SMTP server and close the connection.
	// This interface is particularly useful for unit testing, allowing you to simulate the behavior of the SMTP server's data writer.
	writeCloser interface {
		Write([]byte) (int, error)
		Close() error
	}
)

// default configs where mailer will be configured initially if no specific configuration is passed.
// defaultTLSCfg returns default tls.Config.
func defaultTLSCfg(host string) *tls.Config {
	return &tls.Config{
		ServerName: host,
	}
}

func defaultDialTimeout() time.Duration {
	return time.Second * 5
}

// WithLocalName configures Mailer with localName.
func WithLocalName(l string) func(mailer *Mailer) {
	return func(mailer *Mailer) {
		if l != "" {
			mailer.localName = l
		}
	}
}

// WithTLSConfig configures Mailer with tls.Config.
func WithTLSConfig(cfg *tls.Config) func(*Mailer) {
	return func(mailer *Mailer) {
		if cfg != nil {
			mailer.tlsConfig = cfg
		}
	}
}

// WithDialTimeout configures Mailer with time.Duration for dial timeout.
func WithDialTimeout(t time.Duration) func(*Mailer) {
	return func(mailer *Mailer) {
		if t.Seconds() > 0 {
			mailer.dialTimeout = t
		}
	}
}

// WithAuth configures Mailer with smtp.Auth mechanism.
func WithAuth(auth smtp.Auth) func(*Mailer) {
	return func(mailer *Mailer) {
		if auth != nil {
			mailer.auth = auth
		}
	}
}

// WithSecrets configures Mailer with secrets to authenticate for CRAM-MD5.
func WithSecrets(s string) func(*Mailer) {
	return func(mailer *Mailer) {
		if s != "" {
			mailer.secrets = s
		}
	}
}

// WithSSLEnabled configures Mailer with ssl option.
func WithSSLEnabled(s bool) func(*Mailer) {
	return func(mailer *Mailer) {
		if s {
			mailer.sslEnabled = s
		}
	}
}

// Mailer encapsulates the connection overhead and holds the email functionality.
// It provides methods to send emails with and without TLS.
type Mailer struct {
	// Port represents the port of the SMTP server.
	Port int
	// Host represents the host of the SMTP server.
	Host string
	// Username is used to authenticate to the SMTP server.
	Username string
	// Password is the password to use to authenticate to the SMTP server.
	Password string
	// localName is the hostname sent to the SMTP server.
	localName string
	// auth represents the way of authentication to a given SMTP server.
	auth smtp.Auth
	// tlsConfig represents the TLS configuration used.
	tlsConfig *tls.Config

	// sslEnabled indicates whether SSL is enabled.
	sslEnabled bool

	// secrets used for CRAM-MD5 authentication.
	secrets string

	// dialTimeout represents a timeout configuration for connecting to smtp server.
	dialTimeout time.Duration
}

// NewMailer creates a new mailer to send emails via smtp.
func NewMailer(host string, port int, username, password string, opts ...Options) *Mailer {
	mailer := &Mailer{
		Port:        port,
		Username:    username,
		Password:    password,
		Host:        host,
		tlsConfig:   defaultTLSCfg(host),
		dialTimeout: defaultDialTimeout(),
	}
	if opts != nil {
		// Applying options.
		for _, opt := range opts {
			opt(mailer)
		}
	}
	return mailer
}

// ConnectAndAuthenticate connects and authenticates the Mailer to an SMTP server and saves the connection internally.
// To terminate the connection, the consumer must issue a Mailer.Close call after they finish sending emails.
//
// Returns:
//
//	SendCloser: An interface that provides methods to send emails and close the connection.
//	error: An error if the connection or authentication fails, or nil if successful.
//
// The function performs the following steps:
// 1. Establishes a TLS connection to the SMTP server using the provided host and port.
// 2. If SSL is enabled (port is 465), it wraps the connection with TLS.
// 3. Creates a new SMTP client using the established connection.
// 4. If a local name is provided, it sends a HELO/EHLO command with the local name.
// 5. If the port is not 465, it checks for the STARTTLS extension and starts TLS if supported.
// 6. Checks for supported authentication mechanisms and sets the appropriate authentication method.
// 7. Authenticates with the SMTP server using the selected authentication method.
// 8. Returns a mailSender instance that implements the SendCloser interface.
func (m *Mailer) ConnectAndAuthenticate() (SendCloser, error) {
	netConn, err := netDialTimeout("tcp", m.addr(), m.dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to dial to smtp server: %w", err)
	}
	// check if ssl is enabled.
	if m.Port == sslPort {
		netConn = tlsClient(netConn, m.tlsConfig)
	}
	c, err := newSmtpClient(netConn, m.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to dial smtp server: %w", err)
	}
	if m.localName != "" {
		if err := c.Hello(m.localName); err != nil {
			return nil, fmt.Errorf("failed to dial smtp server: %w", err)
		}
	}

	if !m.sslEnabled {
		// check if conn starts with tls
		// if starts apply tls config.
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(m.tlsConfig); err != nil {
				c.Close()
				return nil, fmt.Errorf("failed to StartTLS: %w", err)
			}
		}
	}
	// check if auth is given or determine which auth mechanism to use.
	if m.auth == nil && m.Username != "" {
		m.authenticationMechanism(c)
	}
	// authenticate
	if m.auth != nil {
		if err = c.Auth(m.auth); err != nil {
			c.Close()
			return nil, fmt.Errorf("failed to authenticate with smtp server: %w", err)
		}
	}
	return &mailSender{m, c}, nil
}

// authenticationMechanism function set the authentication mechanism for smtp server.
func (m *Mailer) authenticationMechanism(smtpClient smtpClient) {
	if ok, auths := smtpClient.Extension("AUTH"); ok {
		if strings.Contains(auths, crmAuthMechanism) {
			m.auth = smtpCRAMMD5Auth(m.Username, m.secrets)
		} else if strings.Contains(auths, plainAuthMechanism) {
			m.auth = smtpPlainAuth("", m.Username, m.Password, m.Host)
		} else {
			m.auth = newSmtpLoginAuth(m.Username, m.Password)
		}
	}
}

// Send dials the SMTP server with the proper authentication and sends an email.
//
// Parameters:
//
//   - message (message.Message): The message to be sent.
//
// Returns:
//
//   - error: An error if the email could not be sent, or nil if the email was sent successfully.
//
// The function performs the following steps:
// 1. Connects and authenticates to the SMTP server using the `ConnectAndAuthenticate` method of the `Mailer` struct.
// 2. Sends the email using the `Send` method of the `SendCloser` interface.
// 3. Closes the connection to the SMTP server.
//
// Example usage:
//
//	mailer := NewMailer("smtp.example.com", 465, "user@example.com", "password")
//	message := message.Message{
//	    From:       "sender@example.com",
//	    Recipients: []string{"recipient@example.com"},
//	    Body:       "This is a test email.",
//	}
//	err := mailer.Send(message)
//	if err != nil {
//	    log.Fatalf("Failed to send email: %v", err)
//	}
func (m *Mailer) Send(message message.Message) error {
	sender, err := m.ConnectAndAuthenticate()
	if err != nil {
		return fmt.Errorf("failed to connect and authenticate: %w", err)
	}
	defer sender.Close()

	if err := sender.Send(message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// addr returns full adders.
func (m *Mailer) addr() string {
	return fmt.Sprintf("%s:%d", m.Host, m.Port)
}

// mailSender is a data struct that promotes the functionality of smtp.Client and supports features of Mailer.
type mailSender struct {
	// mailer is a reference to the Mailer instance that created this mailSender.
	mailer *Mailer
	// smtpClient is the SMTP client used to send emails.
	smtpClient
}

// Send sends the provided message using the SMTP client.
//
// Parameters:
//   - message (message.Message): The message to be sent.
//
// Returns:
//   - error: An error if the message could not be sent, or nil if the message was sent successfully.
//
// The function performs the following steps:
// 1. Sends the MAIL command with the sender's address.
// 2. Sends the RCPT command for each recipient's address.
// 3. Initiates the DATA command to start the message data transfer.
// 4. Encodes the message and writes it to the SMTP client's data writer.
// 5. Closes the data writer.
//
// If any step fails, an appropriate error is returned.
func (m *mailSender) Send(message message.Message) error {
	if err := m.Mail(message.From); err != nil {
		return fmt.Errorf("mailer failed to send MAIL command for address %s: %w", message.From, err)
	}

	for _, t := range message.Recipients {
		if err := m.Rcpt(t); err != nil {
			return fmt.Errorf("mailer failed to send rcpt command for address %s: %w", t, err)
		}
	}
	w, err := m.Data()
	if err != nil {
		return fmt.Errorf("mailer failed to get data writer: %w", err)
	}
	encodedMsg, err := message.Encode()
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	_, err = w.Write(encodedMsg)
	defer func() {
		_ = w.Close()
	}()
	if err != nil {
		return fmt.Errorf("failed writing data: %w", err)
	}

	return nil
}

// Close closes the connection between the client and the SMTP server.
//
// Returns:
//   - error: An error if the connection could not be closed, or nil if the connection was closed successfully.
//
// The function performs the following steps:
// 1. Sends the QUIT command to the SMTP server to terminate the session.
// 2. If the QUIT command fails, it returns an error indicating the failure.
// 3. If the QUIT command succeeds, it returns nil.
func (m *mailSender) Close() error {
	if err := m.Quit(); err != nil {
		return fmt.Errorf("failed to close connection to smtp server: %w", err)
	}
	return nil
}

// Extracted functions to be stubbed during testing to avoid dialing a real server.
// These functions are used to create mock implementations for unit tests,
// ensuring that the tests do not make actual network connections.
var (
	// newSmtpClient returns smtpClient interface.
	newSmtpClient = func(conn net.Conn, host string) (smtpClient, error) {
		return smtp.NewClient(conn, host)
	}

	// smtpPlainAuth returns smtp.PlainAuth.
	smtpPlainAuth = func(identity, username, password, host string) auth {
		return smtp.PlainAuth(identity, username, password, host)
	}
	// tlsClient returns tlsClient.
	tlsClient = tls.Client

	// smtpCRAMMD5Auth returns smtp.smtpCRAMMD5Auth.
	smtpCRAMMD5Auth = smtp.CRAMMD5Auth
	// netDialTimeout returns net.DialTimeout func.
	netDialTimeout = net.DialTimeout
)
