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
	mailSubjectEncoded := "=?UTF-8?B?" + encodeBase64(m.Subject) + "?="
	headers := make(map[string]string)
	hasAttachement := len(m.Attachments) > 0
	headers["MIME-Version"] = "1.0"

	if hasAttachement {
		headers["Content-Type"] = multiPartContentType
	} else if m.HTMLBody != "" {
		headers["Content-Type"] = htmlTypeContentType
	} else {
		headers["Content-Type"] = defaultContentType
	}
	headers["Subject"] = mailSubjectEncoded
	headers["From"] = m.From

	if len(m.Recipients) > 0 {
		headers["To"] = strings.Join(m.Recipients, separator)
	}
	if len(m.Cc) > 0 {
		headers["Cc"] = strings.Join(m.Cc, separator)
	}

	if len(m.Bcc) > 0 {
		headers["Bcc"] = strings.Join(m.Bcc, separator)
	}

	for k, v := range m.Headers {
		headers[k] = v[0]
	}
	var mailMessage strings.Builder
	for k, v := range headers {
		mailMessage.WriteString(fmt.Sprintf("%s: %s%s", k, v, crlf))
	}
	mailMessage.WriteString(crlf)

	// if Message has attachement
	if hasAttachement {
		mailMessage.WriteString(fmt.Sprintf("--%s%s", boundary, crlf))
		if m.HTMLBody != "" {
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", htmlTypeContentType, crlf))
		} else {
			mailMessage.WriteString(fmt.Sprintf("Content-Type: %s%s", defaultContentType, crlf))
		}
		mailMessage.WriteString(fmt.Sprintf("Content-Transfer-Encoding: 7bit%s", crlf))
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
