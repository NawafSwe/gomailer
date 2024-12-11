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
func encode(m Message) []byte {
	var mailMessage strings.Builder
	mailSubjectEncoded := "=?UTF-8?B?" + encodeBase64(m.Subject) + "?="
	hasAttachement := len(m.Attachments) > 0
	hasBothPlainAndHTML := m.Body != "" && m.HTMLBody != ""
	mailMessage.WriteString(fmt.Sprintf("MIME-Version: 1.0%s", crlf))
	mailMessage.WriteString(fmt.Sprintf("Subject: %s%s", mailSubjectEncoded, crlf))
	mailMessage.WriteString(fmt.Sprintf("From: %s%s", m.From, crlf))
	// set the main content type
	if hasAttachement {
		mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", multiPartMixedContentType, crlf))
	} else if hasBothPlainAndHTML {
		mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", multiPartAlternativeContentType, crlf))
	} else if m.HTMLBody != "" {
		mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", htmlTypeContentType, crlf))
	} else {
		mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", plainContentType, crlf))
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
	mailMessage.WriteString(crlf)

	// if Message has attachement
	if hasAttachement {
		mailMessage.WriteString(fmt.Sprintf("--%s%s", boundary, crlf))
		mailMessage.WriteString(encodeMultiPartMixed(m))
		// Add attachments
		for _, attachment := range m.Attachments {
			mailMessage.WriteString(attachment.encode())
		}
		// Final boundary to indicate the end of the message
		mailMessage.WriteString(fmt.Sprintf("--%s--%s", boundary, crlf))

	} else {
		mailMessage.WriteString(encodeMessageContent(m))
	}
	return []byte(mailMessage.String())
}

// encodeMessageContent function encodes the Message.Body, and Message.HTMLBody.
func encodeMessageContent(m Message) string {
	var mb strings.Builder
	// check if mail has both versions.
	if m.Body != "" && m.HTMLBody != "" {
		mb.WriteString(fmt.Sprintf("Content-Type: %s%s", multiPartAlternativeContentType, crlf))
		mb.WriteString(fmt.Sprintf("--%s%s", altBoundary, crlf))
		// Plain text content.
		mb.WriteString(fmt.Sprintf("Content-Type: %s%s", plainContentType, crlf))
		mb.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
		mb.WriteString(crlf)
		for _, line := range splitLines(m.Body, maxLineLength) {
			mb.WriteString(line + crlf)
		}

		mb.WriteString(crlf)
		// HTML content.

		mb.WriteString(fmt.Sprintf("--%s%s", altBoundary, crlf))
		mb.WriteString(fmt.Sprintf("Content-Type: %s%s", htmlTypeContentType, crlf))
		mb.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
		mb.WriteString(crlf)
		mb.WriteString(m.HTMLBody + crlf)
		// Closing boundary
		mb.WriteString(fmt.Sprintf("--%s--%s", altBoundary, crlf))
	} else if m.HTMLBody != "" {
		mb.WriteString(m.HTMLBody + crlf)
	} else {
		for _, line := range splitLines(m.Body, maxLineLength) {
			mb.WriteString(line + crlf)
		}
	}

	return mb.String()
}

// encodeMultiPartMixed function encodes multipart mixed and encodeMessageContent if any.
func encodeMultiPartMixed(m Message) string {
	var mb strings.Builder
	// check if mail has content as alternative
	if m.HTMLBody != "" && m.Body != "" {
		mb.WriteString(encodeMessageContent(m))
	} else if m.HTMLBody != "" {
		mb.WriteString(fmt.Sprintf("Content-Type: %s%s", htmlTypeContentType, crlf))
		mb.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
		mb.WriteString(crlf)
		mb.WriteString(m.HTMLBody)
	} else {
		mb.WriteString(fmt.Sprintf("Content-Type: %s%s", plainContentType, crlf))
		mb.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
		mb.WriteString(crlf)
		for _, line := range splitLines(m.Body, maxLineLength) {
			mb.WriteString(line + crlf)
		}
	}
	mb.WriteString(crlf)

	return mb.String()
}
