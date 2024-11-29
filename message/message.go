package message

import (
	"fmt"
	"net/mail"
)

const (
	// maxLineLength email content is split into lines that do not exceed the maximum length specified by RFC 2045.
	maxLineLength = 76

	// defaultContentType is the default Content-Type according to RFC 2045, section 5.2
	defaultContentType = "text/plain; charset=us-ascii"
	// htmlTypeContentType to support content type with HTML.
	htmlTypeContentType = "text/html; charset=UTF-8"
	boundary            = "my-boundary-12345"

	// https://datatracker.ietf.org/doc/html/rfc5322
	crlf      = "\r\n"
	separator = ", "
)

var (
	multiPartContentType = fmt.Sprintf("multipart/mixed; boundary=%s", boundary)
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

	// Attachments any files attached to email.
	Attachments []Attachment
}

func NewMessage() Message {
	return Message{
		Attachments: make([]Attachment, 0),
	}
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

// Attachment attached files to Message.
type Attachment struct {
	Filename string
	Data     []byte
	MIMEType string
}
