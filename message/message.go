package message

import (
	"fmt"
	"net/mail"
	"strings"
)

const (
	// maxLineLength email content is split into lines that do not exceed the maximum length specified by RFC 2045.
	maxLineLength = 76

	// plainContentType is the default Content-Type according to RFC 2045, section 5.2
	plainContentType = "text/plain; charset=us-ascii"
	// htmlTypeContentType to support content type with HTML.
	htmlTypeContentType = "text/html; charset=UTF-8"

	// The boundary string is used to separate different parts of a multipart email message.
	// This is essential for correctly formatting emails with attachments or multiple content types.
	// For more details, refer to: https://datatracker.ietf.org/doc/html/rfc2046
	boundary = "BOUNDARY"

	// The altBoundary string is used to separate different parts of a multipart email message.
	// This is essential for correctly formatting emails with attachments or multiple content types.
	// For more details, refer to: https://datatracker.ietf.org/doc/html/rfc2046
	altBoundary = "ALT-BOUNDARY"

	// The crlf sequence is used to terminate lines in email messages, as specified by RFC 5322.
	// This ensures proper formatting and compatibility with email clients and servers.
	// For more details, refer to: https://datatracker.ietf.org/doc/html/rfc5322
	crlf = "\r\n"

	separator = ", "
)

var (
	multiPartMixedContentType       = fmt.Sprintf("multipart/mixed; boundary=%s", boundary)
	multiPartAlternativeContentType = fmt.Sprintf("multipart/alternative; boundary=%s", altBoundary)
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
	// Body and HTMLBody represent the content of the email. If both are present, the email will be sent as multipart/alternative,
	// allowing email clients to choose the most suitable version to display. Ensure that the content of Body and HTMLBody is equivalent
	// to provide a consistent user experience. For more details, refer to: https://datatracker.ietf.org/doc/html/rfc2046#section-5.1.4
	Body, HTMLBody string
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
		return fmt.Errorf("recipients cannot be empty slice")
	}

	for _, r := range m.Recipients {
		if _, err := mail.ParseAddress(r); err != nil {
			return fmt.Errorf("given %s is invalid recipient email: %w", r, err)
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

// encode encodes an attachment in base64 and returns the encoded string.
func (a Attachment) encode() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--%s%s", boundary, crlf))
	sb.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"%s", a.MIMEType, a.Filename, crlf))

	// This header specifies how the attachment's data is encoded for transmission, ensuring that the client can correctly decode and display the file.
	// According to RFC 2045, this is crucial for proper email attachment handling.
	// For more details, refer to: https://datatracker.ietf.org/doc/html/rfc2045
	sb.WriteString(fmt.Sprintf("Content-Transfer-Encoding: base64%s", crlf))
	// Email clients needs this header to be able to render the file as attachement and display proper name when user downloading that attachement.
	// see https://datatracker.ietf.org/doc/html/rfc2183
	sb.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"%s", a.Filename, crlf))
	sb.WriteString(crlf)

	// Encode and wrap in 76-char lines
	base64Encoded := encodeBase64(string(a.Data))
	for _, line := range splitLines(base64Encoded, maxLineLength) {
		sb.WriteString(line + crlf)
	}

	sb.WriteString(crlf)
	return sb.String()
}
