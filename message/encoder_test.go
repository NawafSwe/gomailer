package message

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testEmail = "test.usr@smtp.com"
)

func TestMessage_EncodeBase64(t *testing.T) {
	t.Parallel()
	t.Run("should encode message to base64", func(t *testing.T) {
		t.Parallel()
		expected := "aW5wdXQ"
		msg := encodeBase64("input")
		assert.Equal(t, expected, msg)
	})
}

func TestMessage_SplitLines(t *testing.T) {
	input := "input"
	t.Run("should put input into multiple lines when it is exceeding the max length", func(t *testing.T) {
		lines := splitLines(input, 1)
		assert.Equal(t, len(lines), len(input))
		// assert each line.
		for i, line := range lines {
			assert.Equal(t, input[i:i+1], line)
		}
	})

	t.Run("should put input in one line when it is not exceeding the max length", func(t *testing.T) {
		lines := splitLines(input, 20)
		assert.Equal(t, len(lines), 1)
		assert.Equal(t, lines[0], input)
	})
}

func TestMessage_Encode(t *testing.T) {
	tests := map[string]struct {
		input Message
		want  string
	}{
		"should encode message in the expected format when message has an html only with to,cc, and bcc": {
			input: Message{
				From:       "gomailer@smtp.com",
				Recipients: []string{testEmail},
				Cc:         []string{testEmail},
				Bcc:        []string{testEmail},
				HTMLBody:   "<p>hello</p>",
				Subject:    "testing html body",
			},
			want: "MIME-Version: 1.0\r\nSubject: =?UTF-8?B?dGVzdGluZyBodG1sIGJvZHk?=\r\nFrom: gomailer@smtp.com\r\nContent-Type: text/html; charset=UTF-8\r\nTo: test.usr@smtp.com\r\nCc: test.usr@smtp.com\r\nBcc: test.usr@smtp.com\r\n\r\n<p>hello</p>\r\n",
		},
		"should encode message correctly with both HTML and plain text bodies, including to, cc, and bcc fields": {
			input: Message{
				From:       "gomailer@smtp.com",
				Recipients: []string{testEmail},
				Cc:         []string{testEmail},
				Bcc:        []string{testEmail},
				HTMLBody:   "<p>hello</p>",
				Body:       "hello",
				Subject:    "testing html body",
			},
			want: "MIME-Version: 1.0\r\nSubject: =?UTF-8?B?dGVzdGluZyBodG1sIGJvZHk?=\r\nFrom: gomailer@smtp.com\r\nContent-Type: multipart/alternative; boundary=ALT-BOUNDARY\r\nTo: test.usr@smtp.com\r\nCc: test.usr@smtp.com\r\nBcc: test.usr@smtp.com\r\n\r\nContent-Type: multipart/alternative; boundary=ALT-BOUNDARY\r\n--ALT-BOUNDARY\r\nContent-Type: text/plain; charset=us-ascii\r\nContent-Transfer-Encoding: 8bit\r\n\r\nhello\r\n\r\n--ALT-BOUNDARY\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n<p>hello</p>\r\n--ALT-BOUNDARY--\r\n",
		},
		"should encode message in the expected format when message has an text body only with to,cc, and bcc": {
			input: Message{
				From:       "gomailer@smtp.com",
				Recipients: []string{testEmail},
				Cc:         []string{testEmail},
				Bcc:        []string{testEmail},
				Body:       "hello",
				Subject:    "testing txt body",
			},
			want: "MIME-Version: 1.0\r\nSubject: =?UTF-8?B?dGVzdGluZyB0eHQgYm9keQ?=\r\nFrom: gomailer@smtp.com\r\nContent-Type: text/plain; charset=us-ascii\r\nTo: test.usr@smtp.com\r\nCc: test.usr@smtp.com\r\nBcc: test.usr@smtp.com\r\n\r\nhello\r\n",
		},
		"should encode message correctly with plain text body and attachments, including to, cc, and bcc fields": {
			input: Message{
				From:       "gomailer@smtp.com",
				Recipients: []string{testEmail},
				Cc:         []string{testEmail},
				Bcc:        []string{testEmail},
				Body:       "hello",
				Attachments: []Attachment{{
					Filename: "f1",
					Data:     []byte("byte str"),
					MIMEType: "application/pdf",
				}},
				Subject: "testing txt body with attachment",
			},
			want: "MIME-Version: 1.0\r\nSubject: =?UTF-8?B?dGVzdGluZyB0eHQgYm9keSB3aXRoIGF0dGFjaG1lbnQ?=\r\nFrom: gomailer@smtp.com\r\nContent-Type: multipart/mixed; boundary=BOUNDARY\r\nTo: test.usr@smtp.com\r\nCc: test.usr@smtp.com\r\nBcc: test.usr@smtp.com\r\n\r\n--BOUNDARY\r\nContent-Type: text/plain; charset=us-ascii\r\nContent-Transfer-Encoding: 8bit\r\n\r\nhello\r\n\r\n--BOUNDARY\r\nContent-Type: application/pdf; name=\"f1\"\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment; filename=\"f1\"\r\n\r\nYnl0ZSBzdHI\r\n\r\n--BOUNDARY--\r\n",
		},
		"should encode message correctly with plain text and HTML bodies, including attachments, to, cc, and bcc fields": {
			input: Message{
				From:       "gomailer@smtp.com",
				Recipients: []string{testEmail},
				Cc:         []string{testEmail},
				Bcc:        []string{testEmail},
				Body:       "hello",
				HTMLBody:   "<p>hello</p>",
				Attachments: []Attachment{{
					Filename: "f1",
					Data:     []byte("byte str"),
					MIMEType: "application/pdf",
				}},
				Subject: "testing txt body with attachment",
			},
			want: "MIME-Version: 1.0\r\nSubject: =?UTF-8?B?dGVzdGluZyB0eHQgYm9keSB3aXRoIGF0dGFjaG1lbnQ?=\r\nFrom: gomailer@smtp.com\r\nContent-Type: multipart/mixed; boundary=BOUNDARY\r\nTo: test.usr@smtp.com\r\nCc: test.usr@smtp.com\r\nBcc: test.usr@smtp.com\r\n\r\n--BOUNDARY\r\nContent-Type: multipart/alternative; boundary=ALT-BOUNDARY\r\n--ALT-BOUNDARY\r\nContent-Type: text/plain; charset=us-ascii\r\nContent-Transfer-Encoding: 8bit\r\n\r\nhello\r\n\r\n--ALT-BOUNDARY\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n<p>hello</p>\r\n--ALT-BOUNDARY--\r\n\r\n--BOUNDARY\r\nContent-Type: application/pdf; name=\"f1\"\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment; filename=\"f1\"\r\n\r\nYnl0ZSBzdHI\r\n\r\n--BOUNDARY--\r\n",
		},
		"should encode message in the expected format when message has an html body and attachments with to,cc, and bcc": {
			input: Message{
				From:       "gomailer@smtp.com",
				Recipients: []string{testEmail},
				Cc:         []string{testEmail},
				Bcc:        []string{testEmail},
				HTMLBody:   "<p>hello</p>",
				Attachments: []Attachment{{
					Filename: "f1",
					Data:     []byte("byte str"),
					MIMEType: "application/pdf",
				}},
				Subject: "testing txt body with attachment",
			},
			want: "MIME-Version: 1.0\r\nSubject: =?UTF-8?B?dGVzdGluZyB0eHQgYm9keSB3aXRoIGF0dGFjaG1lbnQ?=\r\nFrom: gomailer@smtp.com\r\nContent-Type: multipart/mixed; boundary=BOUNDARY\r\nTo: test.usr@smtp.com\r\nCc: test.usr@smtp.com\r\nBcc: test.usr@smtp.com\r\n\r\n--BOUNDARY\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n<p>hello</p>\r\n--BOUNDARY\r\nContent-Type: application/pdf; name=\"f1\"\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment; filename=\"f1\"\r\n\r\nYnl0ZSBzdHI\r\n\r\n--BOUNDARY--\r\n",
		},
		"should encode message in the expected format when message has an html body and attachments with to,cc, and bcc and additional headers": {
			input: Message{
				From:       "gomailer@smtp.com",
				Recipients: []string{testEmail},
				Cc:         []string{testEmail},
				Bcc:        []string{testEmail},
				Headers:    map[string][]string{"message-id": {"124"}},
				HTMLBody:   "<p>hello</p>",
				Attachments: []Attachment{{
					Filename: "f1",
					Data:     []byte("byte str"),
					MIMEType: "application/pdf",
				}},
				Subject: "testing txt body with attachment",
			},
			want: "MIME-Version: 1.0\r\nSubject: =?UTF-8?B?dGVzdGluZyB0eHQgYm9keSB3aXRoIGF0dGFjaG1lbnQ?=\r\nFrom: gomailer@smtp.com\r\nContent-Type: multipart/mixed; boundary=BOUNDARY\r\nTo: test.usr@smtp.com\r\nCc: test.usr@smtp.com\r\nBcc: test.usr@smtp.com\r\nmessage-id: 124\r\n\r\n--BOUNDARY\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n<p>hello</p>\r\n--BOUNDARY\r\nContent-Type: application/pdf; name=\"f1\"\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment; filename=\"f1\"\r\n\r\nYnl0ZSBzdHI\r\n\r\n--BOUNDARY--\r\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := string(encode(tc.input))
			fmt.Println(got)
			assert.Equal(t, tc.want, got)
		})
	}
}
