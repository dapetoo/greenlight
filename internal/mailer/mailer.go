package mailer

import (
	"bytes"
	"embed"
	"github.com/go-mail/mail/v2"
	"html/template"
	"log"
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

func (m *Mailer) Send(recipient, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		//return err
		log.Println(err)
		//panic(err)
	}

	//Execute the named template "subject" passing in the dynamic data and storing the result in a bytes.Buffer variable
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	//Follow the same pattern to execute the plainBody template and store the result in the plainbody variable
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	//Follow the same pattern to execute the htmlBody template and store the result in the plainbody variable
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "plainBody", data)
	if err != nil {
		return err
	}

	//mail.NewMessage function to initialize a new mail.Message instance
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	//DialAndSend() opens a connection to the SMTP server, sends the message, then close the connection
	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
