package message

import (
	"encoding/base64"
	"fmt"
	"strings"
)

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

// encode encodes mail components into bytes to be sent.
// TODO: handle alternative use case
// TODO: when mail should have two parts one plain text and one is html.
func encode(m Message) []byte {
	var mailMessage strings.Builder
	mailSubjectEncoded := "=?UTF-8?B?" + encodeBase64(m.Subject) + "?="
	hasAttachement := len(m.Attachments) > 0
	mailMessage.WriteString(fmt.Sprintf("MIME-Version: 1.0%s", crlf))
	mailMessage.WriteString(fmt.Sprintf("Subject: %s%s", mailSubjectEncoded, crlf))
	mailMessage.WriteString(fmt.Sprintf("From: %s%s", m.From, crlf))
	if hasAttachement {
		mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", multiPartContentType, crlf))
	} else if m.HTMLBody != "" {
		mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", htmlTypeContentType, crlf))
	} else {
		mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", defaultContentType, crlf))
	}

	if len(m.Recipients) > 0 {
		mailMessage.WriteString(fmt.Sprintf("To: %s%s", strings.Join(m.Recipients, separator), crlf))
	}
	if len(m.Cc) > 0 {
		mailMessage.WriteString(fmt.Sprintf("Cc: %s%s", strings.Join(m.Cc, separator), crlf))
	}

	if len(m.Bcc) > 0 {
		mailMessage.WriteString(fmt.Sprintf("Bcc: %s%s", strings.Join(m.Bcc, separator), crlf))
	}
	// additional headers if any.
	for k, v := range m.Headers {
		mailMessage.WriteString(fmt.Sprintf("%s: %s%s", k, strings.Join(v, ", "), crlf))
	}
	// How to guarantee order?

	mailMessage.WriteString(crlf)

	// if Message has attachement
	if hasAttachement {
		mailMessage.WriteString(fmt.Sprintf("--%s%s", boundary, crlf))
		if m.HTMLBody != "" {
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", htmlTypeContentType, crlf))
		} else {
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", defaultContentType, crlf))
		}
		mailMessage.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
		mailMessage.WriteString(crlf)
		if m.HTMLBody != "" {
			mailMessage.WriteString(m.HTMLBody)
		} else {
			for _, line := range splitLines(m.Body, maxLineLength) {
				mailMessage.WriteString(line + crlf)
			}
		}
		mailMessage.WriteString(crlf)
	} else {
		if m.HTMLBody != "" {
			mailMessage.WriteString(m.HTMLBody)
		} else {
			for _, line := range splitLines(m.Body, maxLineLength) {
				mailMessage.WriteString(line + crlf)
			}
		}
		return []byte(mailMessage.String())
	}
	// Add attachments
	for _, attachment := range m.Attachments {
		mailMessage.WriteString(attachment.encode())
	}

	// Final boundary to indicate the end of the message
	mailMessage.WriteString(fmt.Sprintf("--%s--%s", boundary, crlf))
	return []byte(mailMessage.String())
}
