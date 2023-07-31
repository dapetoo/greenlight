package mailer

import (
	"embed"
	mail "github.com/go-mail/mail/v2"
	"time"
	//mail "github.com/xhit/go-simple-mail/v2"
)

// Embedded File system to hold email templates
var templateFS embed.FS

//Mailer struct which contains a mail.Dialer

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	//Return a mailer instance containing the dialer and sender information
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}
