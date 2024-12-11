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
		if hasBothPlainAndHTML {
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", multiPartAlternativeContentType, crlf))
			mailMessage.WriteString(fmt.Sprintf("--%s%s", altBoundary, crlf))
			// plain text content.
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", plainContentType, crlf))
			mailMessage.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
			mailMessage.WriteString(crlf)
			for _, line := range splitLines(m.Body, maxLineLength) {
				mailMessage.WriteString(line + crlf)
			}

			mailMessage.WriteString(crlf)

			// html content.
			mailMessage.WriteString(fmt.Sprintf("--%s%s", altBoundary, crlf))
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", htmlTypeContentType, crlf))
			mailMessage.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
			mailMessage.WriteString(crlf)
			mailMessage.WriteString(m.HTMLBody + crlf)
			// closing boundary
			mailMessage.WriteString(fmt.Sprintf("--%s--%s", altBoundary, crlf))

		} else if m.HTMLBody != "" {
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", htmlTypeContentType, crlf))
			mailMessage.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
			mailMessage.WriteString(crlf)
			mailMessage.WriteString(m.HTMLBody)
		} else {
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", plainContentType, crlf))
			mailMessage.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
			mailMessage.WriteString(crlf)
			for _, line := range splitLines(m.Body, maxLineLength) {
				mailMessage.WriteString(line + crlf)
			}
		}
		mailMessage.WriteString(crlf)
	} else {
		if hasBothPlainAndHTML {
			mailMessage.WriteString(fmt.Sprintf("--%s%s", altBoundary, crlf))
			// plain text content.
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", plainContentType, crlf))
			mailMessage.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
			mailMessage.WriteString(crlf)
			for _, line := range splitLines(m.Body, maxLineLength) {
				mailMessage.WriteString(line + crlf)
			}

			mailMessage.WriteString(crlf)

			// html content.
			mailMessage.WriteString(fmt.Sprintf("--%s%s", altBoundary, crlf))
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", htmlTypeContentType, crlf))
			mailMessage.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 8bit%s", crlf))
			mailMessage.WriteString(crlf)
			mailMessage.WriteString(m.HTMLBody + crlf)
			// closing boundary
			mailMessage.WriteString(fmt.Sprintf("--%s--%s", altBoundary, crlf))

		}
		if m.HTMLBody != "" {
			mailMessage.WriteString(m.HTMLBody + crlf)
		} else {
			mailMessage.WriteString(crlf)
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
