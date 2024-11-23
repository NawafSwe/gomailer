package message

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
)

const (
	// maxLineLength email content is split into lines that do not exceed the maximum length specified by RFC 2045.
	maxLineLength = 76

	// defaultContentType is the default Content-Type according to RFC 2045, section 5.2
	defaultContentType = "text/plain; charset=us-ascii"
	// htmlTypeContentType to support content type with HTML.
	htmlTypeContentType = "text/html; charset=UTF-8"
	// message Lint: http://tools.ietf.org/tools/msglint/
	crlf      = "\r\n"
	separator = ", "
)

// Message will be sent in email.
type Message struct {
	// From whom is going to send that mail.
	From string
	// Recipients contains the primary recipients of the email.
	Recipients []string
	// Cc contains the recipients who will receive a carbon copy of the email.
	Cc []string
	// Bcc contains the recipients who will receive a blind carbon copy of the email.
	Bcc []string
	// Body the content of the email.
	Body string
	// HTMLBody the content of the email as HTML.
	HTMLBody string
	// Subject the subject of the email.
	Subject string
	// Headers Extra mail headers
	Headers mail.Header
}

func NewMessage() Message {
	return Message{}
}

// validate validates message primary fields before send operation.
func (m Message) validate() error {
	if m.From == "" {
		return fmt.Errorf("from address cannot be empty")
	}
	if _, err := mail.ParseAddress(m.From); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}
	if len(m.Recipients) == 0 {
		return fmt.Errorf("to cannot be empty")
	}

	for _, r := range m.Recipients {
		if _, err := mail.ParseAddress(r); err != nil {
			return fmt.Errorf("invalid recipient email: %w", err)
		}
	}
	return nil
}
func (m Message) Encode() ([]byte, error) {
	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("failed to encode message: %w", err)
	}
	return encode(m), nil
}

// Send sends email using smtp.Auth.
func Send(m Message, addr string, a smtp.Auth) error {
	if a == nil {
		return fmt.Errorf("smtp.auth cannot be nil")
	}
	// validating email.
	err := m.validate()
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return smtp.SendMail(addr, a, m.From, m.Recipients, encode(m))
}

// SendWithTLS sends email over a tls with an optional tls.Config
// TLS helps establish a secure and trusted connection between the client and server,
// which is essential for applications that handle sensitive data, such as online banking, email, and e-commerce.
func SendWithTLS(e Message, addr string, a smtp.Auth, tlsCfg *tls.Config) error {
	if a == nil {
		return fmt.Errorf("smtp.auth cannot be nil")
	}

	// validating email.
	err := e.validate()
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("failed to dail addr %s: %w", addr, err)
	}
	client, err := smtp.NewClient(conn, tlsCfg.ServerName)
	if err != nil {
		return fmt.Errorf("failed to create a smtp client: %w", err)
	}

	if err = client.Auth(a); err != nil {
		return fmt.Errorf("failed to authenticate with smtp server: %w", err)
	}

	if err = client.Mail(e.From); err != nil {
		return fmt.Errorf("smpt client failed to mail from address %s: %w", e.From, err)
	}
	for _, t := range e.Recipients {
		if err := client.Rcpt(t); err != nil {
			return fmt.Errorf("smtp client failed to send rcpt command to server for address %s: %w", t, err)
		}
	}
	w, err := client.Data()

	if err != nil {
		return fmt.Errorf("failed to get data writer from smtp client: %w", err)
	}
	_, _ = w.Write(encode(e))
	defer func() {
		_ = w.Close()
	}()
	return client.Quit()
}

//type Attachment struct {
//	Filename string
//	Data     []byte
//	MIMEType string
//}
//
//type Message struct {
//	// existing fields...
//	Attachments []Attachment
//}
//
//func (m Message) Encode() ([]byte, error) {
//	// Extend this function to handle MIME multipart encoding for attachments.
//}
