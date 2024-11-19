package gomailer

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
)

//go:generate mockgen -source=mailer.go -destination=mock/mailer.go -package=mock
type (

	// Auth is implemented by an SMTP authentication mechanism.
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
)

// Mailer encapsulate the connection overhead and holds the email functionality.
type Mailer struct {
	port       int
	username   string
	password   string
	host       string
	localName  string
	ssl        bool
	tlsCfg     *tls.Config
	smtpClient smtpClient
	auth       auth
}

func defaultTLSCfg(host string) *tls.Config {
	return &tls.Config{
		ServerName: host,
	}
}

// A MailerConfig represents the Mailer config to be used to connect to SMTP server.
type MailerConfig struct {

	// Port represents the port of the SMTP server.
	Port int
	// Host represents the host of the SMTP server.
	Host string
	// Username is used to authenticate to the SMTP server.
	Username string
	// Password is the password to use to authenticate to the SMTP server.
	Password string
	// Auth represents the way of authentication to a given SMTP server.
	Auth smtp.Auth
	// SSL indicator if an SSL connection is used.
	SSL bool
	// TLSConfig represents the TLS configuration used.
	TLSConfig *tls.Config
	// LocalName is the hostname sent to the SMTP server.
	LocalName string
}

// NewMailer creates a new mailer to send emails.
func NewMailer(cfg MailerConfig) (*Mailer, error) {
	// support different auths.
	smtpAuth := cfg.Auth
	if smtpAuth == nil {
		smtpAuth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	// check if tls is enabled.
	var tlsCfg *tls.Config
	if cfg.TLSConfig == nil {
		tlsCfg = defaultTLSCfg(cfg.Host)
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to smtp server: %w", err)
	}
	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to create smtp client: %w", err)
	}
	if err = c.Auth(smtpAuth); err != nil {
		return nil, fmt.Errorf("failed authenticate smtp client: %w", err)
	}
	return &Mailer{
		port:       cfg.Port,
		username:   cfg.Username,
		password:   cfg.Password,
		host:       cfg.Host,
		localName:  cfg.LocalName,
		ssl:        cfg.SSL,
		tlsCfg:     tlsCfg,
		smtpClient: c,
		auth:       smtpAuth,
	}, nil
}
