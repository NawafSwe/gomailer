# GoMailer Library

This Go Mailer Library provides a convenient way to send emails with support for attachments, multipart messages (HTML and plain text), and various authentication mechanisms. It supports features such as:

- Sending emails with plain text, HTML content, or both (multipart/alternative).
- Sending attachments with base64 encoding.
- Customizable headers for the email message.
- Handling different email formats (e.g., text/plain, text/html, multipart/alternative, multipart/mixed).

## Installation

To install the library, run:
```shell
go get github.com/NawafSwe/gomailer
```

# Usage
Creating a Mailer Client
First, you need to create a new mailer client by calling the NewMailer function and passing the necessary parameters such as host, port, username, and password. You can also use the provided configuration options to customize the mailer client.
```go 
package main

import (
    "crypto/tls"
    "time"
	
    "github.com/NawafSwe/gomailer"
)

func main() {
    // Create a new mailer client with basic configuration
    mailer := gomailer.NewMailer(
        "smtp.example.com", // SMTP server host
        587,                // SMTP server port
        "user@example.com", // Username
        "password",         // Password
        // Additional configuration options
        gomailer.WithLocalName("localhost"),
        gomailer.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
        gomailer.WithDialTimeout(10*time.Second),
        gomailer.WithSSLEnabled(true),
    )

    // Use the mailer client to send emails
    // ...
}
```
# Configuration Options
 Here are the available configuration options you can use with the NewMailer function:  
- WithLocalName: Configures the mailer with a local name.
- WithTLSConfig: Configures the mailer with a custom tls.Config.
- WithDialTimeout: Configures the mailer with a custom dial timeout.
- WithAuth: Configures the mailer with a custom SMTP authentication mechanism.
- WithSecrets: Configures the mailer with secrets for CRAM-MD5 authentication.
- WithSSLEnabled: Configures the mailer to use SSL.


# Sending an Email
Once you have configured the mailer client, you can construct and send email messages. Here is how you can do it:
Create a New Message: Instantiate a new Message struct using the NewMessage function.
Set Message Fields: Set the necessary fields such as From, Recipients, Subject, Body, HTMLBody, etc.
Send the Message: Use the Send method of the mailer client to send the message.

```go
package main

import (
    "log"
	
    "github.com/NawafSwe/gomailer"
    "github.com/NawafSwe/gomailer/message"
)

func main() {
    // Create a new mailer client
    mailer := gomailer.NewMailer(
        "smtp.example.com",
        587,
        "user@example.com",
        "password",
        gomailer.WithLocalName("localhost"),
        gomailer.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
        gomailer.WithDialTimeout(10*time.Second),
        gomailer.WithSSLEnabled(true),
    )

    // Create a new email message
    msg := message.NewMessage()
    msg.From = "sender@example.com"
    msg.Recipients = []string{"recipient@example.com"}
    msg.Subject = "Test Email"
    msg.Body = "This is a plain text email body"
    msg.HTMLBody = "<p>This is an <b>HTML</b> email body</p>"

    // Optional: Set Cc, Bcc, or custom headers
    msg.Cc = []string{"cc@example.com"}
    msg.Bcc = []string{"bcc@example.com"}

    // Optional: Add attachments
    attachment := message.Attachment{
        Filename: "document.pdf",
        Data:     []byte("binary-data"),
        MIMEType: "application/pdf",
    }
    msg.Attachments = append(msg.Attachments, attachment)

    // Send the message
    if err := mailer.Send(msg); err != nil {
        log.Fatalf("failed to send email: %v", err)
    }
}
```

# Connecting and Authenticating Once
To avoid establishing a connection to the SMTP server every time you send an email, you can use the ```ConnectAndAuthenticate``` method to connect and authenticate once, and then reuse the connection for multiple emails. Remember to call the ```Close``` method after you finish sending emails to terminate the connection.
```go 
package main

import (
    "log"
    "github.com/NawafSwe/gomailer"
    "github.com/NawafSwe/gomailer/message"
)

func main() {
    // Create a new mailer client
    mailer := gomailer.NewMailer(
        "smtp.example.com",
        587,
        "user@example.com",
        "password",
        gomailer.WithLocalName("localhost"),
        gomailer.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
        gomailer.WithDialTimeout(10*time.Second),
        gomailer.WithSSLEnabled(true),
    )

    // Connect and authenticate once
    sender, err := mailer.ConnectAndAuthenticate()
    if err != nil {
        log.Fatalf("failed to connect and authenticate: %v", err)
    }
    defer sender.Close()

    // Create a new email message
    msg := message.NewMessage()
    msg.From = "sender@example.com"
    msg.Recipients = []string{"recipient@example.com"}
    msg.Subject = "Test Email"
    msg.Body = "This is a plain text email body"
    msg.HTMLBody = "<p>This is an <b>HTML</b> email body</p>"

    // Send the message
    if err := sender.Send(msg); err != nil {
        log.Fatalf("failed to send email: %v", err)
    }
}
```


# Features
- Plain Text and HTML Emails: Send emails with plain text, HTML content, or both.
- Attachments: Attach files to your emails with base64 encoding.
- Custom Headers: Add custom headers to your email messages.
- Multiple Recipients: Support for To, Cc, and Bcc recipients.

# License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.