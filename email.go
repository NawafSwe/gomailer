package gomailer

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
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

// Email will be sent in email.
type Email struct {
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

func NewEmail() Email {
	return Email{}
}

// validate validates email primary fields before send operation.
func (e Email) validate() error {
	if e.From == "" {
		return fmt.Errorf("from address cannot be empty")
	}
	if _, err := mail.ParseAddress(e.From); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}
	if len(e.Recipients) == 0 {
		return fmt.Errorf("to cannot be empty")
	}

	for _, r := range e.Recipients {
		if _, err := mail.ParseAddress(r); err != nil {
			return fmt.Errorf("invalid recipient email: %w", err)
		}
	}
	return nil
}

// Send sends email using smtp.Auth.
func (e Email) Send(addr string, a smtp.Auth) error {
	if a == nil {
		return fmt.Errorf("smtp.auth cannot be nil")
	}
	// validating email.
	err := e.validate()
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return smtp.SendMail(addr, a, e.From, e.Recipients, encodeEmail(e))
}

// SendWithTLS sends email over a tls with an optional tls.Config
// TLS helps establish a secure and trusted connection between the client and server,
// which is essential for applications that handle sensitive data, such as online banking, email, and e-commerce.
func (e Email) SendWithTLS(addr string, a smtp.Auth, tlsCfg *tls.Config) error {
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
	_, _ = w.Write(encodeEmail(e))
	defer func() {
		_ = w.Close()
	}()
	return client.Quit()
}

// encodeBase64 Helper function to encode a string in Base64.
func encodeBase64(input string) string {
	return strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(input)), "=")
}

// splitLines splits the input string into lines of a specified maximum length.
func splitLines(input string, maxLength int) []string {
	var lines []string
	for len(input) > maxLength {
		lines = append(lines, input[:maxLength])
		input = input[maxLength:]
	}
	lines = append(lines, input)
	return lines
}

// encodeEmail encodes mail components into bytes to be sent.
func encodeEmail(e Email) []byte {
	mailSubjectEncoded := "=?UTF-8?B?" + encodeBase64(e.Subject) + "?="
	headers := make(map[string]string)
	headers["MIME-Version"] = "1.0"
	if e.HTMLBody != "" {
		headers["Content-Type"] = htmlTypeContentType
	} else {
		headers["Content-Type"] = defaultContentType
	}
	headers["Subject"] = mailSubjectEncoded
	headers["From"] = e.From

	if len(e.Recipients) > 0 {
		headers["To"] = strings.Join(e.Recipients, separator)
	}
	if len(e.Cc) > 0 {
		headers["Cc"] = strings.Join(e.Cc, separator)
	}

	if len(e.Bcc) > 0 {
		headers["Bcc"] = strings.Join(e.Bcc, separator)
	}

	for k, v := range e.Headers {
		headers[k] = v[0]
	}
	var mailMessage strings.Builder
	for k, v := range headers {
		mailMessage.WriteString(fmt.Sprintf("%s: %s%s", k, v, crlf))
	}
	mailMessage.WriteString(crlf)
	if e.HTMLBody != "" {
		mailMessage.WriteString(e.HTMLBody)
	} else {
		for _, line := range splitLines(e.Body, maxLineLength) {
			mailMessage.WriteString(line + crlf)
		}
	}
	return []byte(mailMessage.String())
}
