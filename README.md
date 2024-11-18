## Gomailer
Gomailer is a library built on top of Go's standard `net/smtp` package with additional features such as support for HTML content, attachments, and connection pools.

## Features

- **Sending Emails**:
  - Support for sending emails with and without HTML content.
  - Ability to send emails to multiple recipients, including CC and BCC.
- **Error Handling**:
  - Comprehensive error handling to ensure robust email sending.
- **Email Components**:
  - Abstraction of main email components such as body, recipients, CC, BCC, and subject.
- **Authentication**:
  - Support for multiple authentication methods, including plain authentication.
- **Attachments**:
  - Support for adding attachments to emails.
- **Connection Pools**:
  - Efficient management of SMTP connections using connection pools.

## Usage

### Sending a Simple Email

```go
package main

import (
	"fmt"
	"github.com/NawafSwe/gomailer"
	"net/smtp"
)

func main() {
	email := gomailer.NewEmail()
	email.From = "sender@example.com"
	email.Recipients = []string{"recipient@example.com"}
	email.Subject = "Test Email"
	email.Body = "This is a test email."

	auth := smtp.PlainAuth("", "username", "password", "smtp.example.com")
	err := email.Send("smtp.example.com:587", auth)
	if err != nil {
		fmt.Println("Failed to send email:", err)
	} else {
		fmt.Println("Email sent successfully!")
	}
}

```

### Sending an Email with HTML Content

```go
package main

import (
  "fmt"
  "github.com/NawafSwe/gomailer"
  "net/smtp"
)

func main() {
    email := gomailer.NewEmail()
    email.From = "sender@example.com"
    email.Recipients = []string{"recipient@example.com"}
    email.Subject = "Test Email with HTML"
    email.HTMLBody = "<h1>This is a test email.</h1>"

    auth := smtp.PlainAuth("", "username", "password", "smtp.example.com")
    err := email.Send("smtp.example.com:587", auth)
    if err != nil {
        fmt.Println("Failed to send email:", err)
    } else {
        fmt.Println("Email sent successfully!")
    }
}
```

## Installation

To install Gomailer, use `go get`:

```sh
go get github.com/NawafSWE/gomailer
```

## TODOs

- [ ] Support multi-part content
- [ ] Implement the actual authentication to use auth/plain auth/OAuth etc. (Building on top of the actual Go dialer and SMTP)

## License

Gomailer is licensed under the MIT license. See the [`LICENSE`](LICENSE) file for more details.