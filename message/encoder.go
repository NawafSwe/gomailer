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
func encode(e Message) []byte {
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
