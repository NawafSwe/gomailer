# Go Email Library

A simple and extensible Go library for sending emails with support for attachments, HTML and plain text bodies, different authentication mechanisms, and various email content types.

## Features

- **Email Content Types:**
  - Supports **plain text** (`text/plain`)
  - Supports **HTML** (`text/html`)
  - Supports **multipart/alternative** for emails with both plain text and HTML content.
  - Supports **multipart/mixed** for emails with attachments.

- **Attachments:**
  - Attachments can be added to emails with **base64 encoding**.
  - Supports any MIME type for attachments (e.g., PDF, images, etc.).
  - Automatically generates headers like `Content-Disposition` and `Content-Transfer-Encoding` for attachments.

- **Validation:**
  - Ensures that the essential fields like "From" and "Recipients" are set and valid.
  - Validates email addresses for compliance with RFC standards.

- **Customization:**
  - Custom headers can be added to the email message.
  - Allows setting **multiple recipients**, including **CC** and **BCC**.

- **Easy-to-Use Interface:**
  - The library exposes simple functions to create and send emails with minimal configuration.
  - Automatically handles the complexities of multipart emails, encoding, and content types.

- **Comprehensive Encoding:**
  - Encodes the email in accordance with **RFC 5322** and **RFC 2045**, ensuring compatibility with email servers and clients.
  - Automatically splits long lines (76 characters) in content to comply with email standards.

- **Supports Different Authentication Mechanisms:**
  - Integrates with SMTP, supporting **plain authentication**, **server authentication**, and **TLS** for secure email transmission.

## Installation

To install this library, use the following command:

```bash
go get github.com/NawafSwe/gomailer
```


### Usage
Create a New Email Message
Create a new email message using NewMessage():

```go

package main

import (
	"fmt"
	"github.com/NawafSwe/gomailer/message"
)

func main() {
	// Create a new message
	email := message.NewMessage()

	// Set email fields
	email.From = "sender@example.com"
	email.Recipients = []string{"recipient@example.com"}
	email.Subject = "Test Email"
	email.Body = "This is a plain text body."
	email.HTMLBody = "<h1>This is an HTML Body</h1>"

	// Add attachments if needed
	email.Attachments = append(email.Attachments, message.Attachment{
      Filename: "test.pdf",
      Data:     []byte("PDF data here"),
      MIMEType: "application/pdf",
    })

	// Send the email (replace with your SMTP configuration)
	encodedEmail, err := email.Encode()
	if err != nil {
		fmt.Println("Error encoding email:", err)
		return
	}

	// Simulate sending email (you would use an SMTP client here)
	fmt.Println(string(encodedEmail))
}
```
Key Methods:
NewMessage(): Creates a new empty email message object.
Message.Encode(): Encodes the email message, validating required fields before encoding.
Message.Validate(): Validates the "From" address and "Recipients" email addresses.
Attachment.Encode(): Encodes an email attachment in base64 format and generates necessary headers.
Attachments Support
You can attach files to emails by adding Attachment objects to the Attachments field. Each attachment needs a filename, data, and MIME type.

go
Copy code
email.Attachments = append(email.Attachments, message.Attachment{
	Filename: "file.txt",
	Data:     []byte("Hello, world!"),
	MIMEType: "text/plain",
})
HTML and Plain Text Body Support
You can provide both HTML and plain text bodies for the email. The library will handle sending the appropriate part in a multipart/alternative format.

go
Copy code
email.Body = "This is a plain text body"
email.HTMLBody = "<h1>This is an HTML Body</h1>"
If only one body is provided, the library will send it in the appropriate format (text/plain or text/html).

Custom Headers
You can add custom headers to the email by using the Headers field.

go
Copy code
email.Headers = message.mail.Header{
	"X-Custom-Header": []string{"Custom Value"},
}
Sending the Email
Once the message is fully populated, use the Encode() method to get the email's byte representation, which can be sent via your SMTP client.

Why It's Easy to Use
Simple API: The library is designed with simplicity in mind. You don't need to worry about the complex internals of MIME encoding or boundary management. You can create and send emails by filling in just a few fields.

Automatic MIME Handling: The library automatically determines the correct content type for the email (plain text, HTML, or multipart) and handles the necessary encoding and content separation, making it easy to send rich emails.

Attachment Support: Attachments are as simple as adding a Attachment object to the Attachments slice. The library takes care of the encoding, header generation, and formatting.

Validation: The built-in validation ensures that all necessary fields are set correctly before the email is encoded, reducing the chances of sending incomplete or invalid emails.

Extensible: You can easily extend or modify the library to add more features like richer header handling, or implement advanced SMTP authentication.

License
This library is licensed under the MIT License - see the LICENSE file for details.

markdown
Copy code

### Key Points Covered:

- **Features**: List of features of the library (content types, attachments, validation, customization).
- **Installation**: How to install and use the library.
- **Usage**: Example code for creating and sending an email.
- **Why it's easy to use**: Highlights the simplicity and extensibility of the library, focusing on its ease of integration and automatic handling of email-related complexities.

This README provides a comprehensive overview and clear instructions for users to get started with the library.
